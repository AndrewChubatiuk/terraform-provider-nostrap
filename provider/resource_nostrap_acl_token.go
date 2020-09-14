package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/schema"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	WAIT_TIMEOUT             = 5
	NOMAD_BOOTSTRAP_ENDPOINT = "/v1/acl/bootstrap"
	NOMAD_NODES_ENDPOINT = "/v1/nodes"
)

type NostrapAclTokenResponse struct {
	AccessorID  string
	SecretID    string
	Name        string
	Type        string
	Policies    string
	Global      bool
	CreateTime  string
	CreateIndex int
	ModifyIndex int
}

func resourceNostrapAclToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceNostrapAclTokenCreate,
		Read:   resourceNostrapAclTokenRead,
		Update: resourceNostrapAclTokenUpdate,
		Delete: resourceNostrapAclTokenDelete,
		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nomad cluster address",
				ForceNew:    true,
			},
			"ssm_prefix": &schema.Schema{
				Type:        schema.TypeString,
				Description: "AWS SSM Prefix",
				Required:    true,
			},
			"accessor_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Nomad Token Accessor ID",
				Computed:    true,
			},
			"secret_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Nomad Token Secret ID",
				Computed:    true,
			},
			"aws_region": &schema.Schema{
				Type:        schema.TypeString,
				Description: "AWS SSM Region",
				Required:    true,
			},
		},
	}
}

func getSsmClient(region string) *ssm.SSM {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil
	}
	return ssm.New(sess)
}

func resourceNostrapAclTokenCreate(d *schema.ResourceData, m interface{}) error {
	var token NostrapAclTokenResponse
	nomadAddress := d.Get("address").(string)
	nostrapEndpoint := nomadAddress + NOMAD_BOOTSTRAP_ENDPOINT
	region := d.Get("aws_region").(string)

	// Validate Nomad address
	url, err := url.ParseRequestURI(nomadAddress)
	if err != nil {
		return err
	}
	nomadSocket := url.Host
	if url.Port() == "" {
		switch url.Scheme {
		case "https":
			nomadSocket = nomadSocket + ":443"
		default:
			nomadSocket = nomadSocket + ":80"
		}
	}

	// Waiting for service to become ready
	for {
		conn, err := net.Dial("tcp", nomadSocket)
		if err != nil {
			time.Sleep(WAIT_TIMEOUT * time.Second)
		} else {
			defer conn.Close()
			break
		}
	}

	// Nomad ACL Bootstrap
	for {
		resp, err := http.Post(nostrapEndpoint, "application/json", nil)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			defer resp.Body.Close()
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&token)
			if err != nil {
				return err
			}
			break
		} else {
			if resp.Body != nil {
				defer resp.Body.Close()
				buffer := new(bytes.Buffer)
				buffer.ReadFrom(resp.Body)
				body := buffer.String()
				if strings.Contains(body, "ACL bootstrap already done") {
					return fmt.Errorf("Nomad ACL Bootstrap was already done")
				}
			}
		}
		time.Sleep(WAIT_TIMEOUT * time.Second)
	}

	credentials := map[string]string{
		"accessor_id": token.AccessorID,
		"secret_id":   token.SecretID,
	}
	client := getSsmClient(region)
	for key, value := range credentials {
		param := fmt.Sprintf("%s/%s", d.Get("ssm_prefix").(string), key)
		_, err := client.PutParameter(&ssm.PutParameterInput{
			Name:      aws.String(param),
			Overwrite: aws.Bool(true),
			Type:      aws.String("SecureString"),
			Value:     aws.String(value),
		})
		if err != nil {
			return err
		}
	}
	return resourceNostrapAclTokenRead(d, m)
}

func resourceNostrapAclTokenRead(d *schema.ResourceData, m interface{}) error {
	region := d.Get("aws_region").(string)
	nomadAddress := d.Get("address").(string)
	nodesEndpoint := nomadAddress + NOMAD_NODES_ENDPOINT
	client := getSsmClient(region)
	for _, key := range []string{"accessor_id", "secret_id"} {
		param := fmt.Sprintf("%s/%s", d.Get("ssm_prefix").(string), key)
		resp, err := client.GetParameter(&ssm.GetParameterInput{
			Name:           aws.String(param),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			if err.(awserr.Error).Code() == ssm.ErrCodeParameterNotFound {
				d.SetId("")
				return nil
			}
			return err
		}
		d.Set(key, *resp.Parameter.Value)
		if key == "accessor_id" {
			d.SetId(*resp.Parameter.Value)
		}
	}
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", nodesEndpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Nomad-Token", d.Get("secret_id").(string))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Cannot get data from cluster with ACL token. Either token is invalid or cluster has problems")
	}
	return nil
}

func resourceNostrapAclTokenUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceNostrapAclTokenDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
