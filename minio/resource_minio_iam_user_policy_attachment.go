package minio

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMinioIAMUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: minioCreateUserPolicyAttachment,
		Read:   minioReadUserPolicyAttachment,
		Delete: minioDeleteUserPolicyAttachment,
		Importer: &schema.ResourceImporter{
			State: minioImportUserPolicyAttachment,
		},
		Schema: map[string]*schema.Schema{
			"policy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateIAMNamePolicy,
			},
			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateMinioIamUserName,
			},
		},
	}
}

func minioCreateUserPolicyAttachment(d *schema.ResourceData, meta interface{}) error {

	var userName = d.Get("user_name").(string)
	var policyName = d.Get("policy_name").(string)
	minioAdmin := meta.(*S3MinioClient).S3Admin
	err := minioAdmin.SetPolicy(policyName, userName, false)
	if err != nil {
		return NewResourceError("Unable to Set User policy", userName+" "+policyName, err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", userName)))

	return minioReadUserPolicyAttachment(d, meta)
}

func minioReadUserPolicyAttachment(d *schema.ResourceData, meta interface{}) error {
	minioAdmin := meta.(*S3MinioClient).S3Admin
	var userName = d.Get("user_name").(string)

	userInfo, errUser := minioAdmin.GetUserInfo(userName)
	if errUser != nil {
		return NewResourceError("Fail to load user Infos", userName, errUser)
	}

	if userInfo.PolicyName == "" {
		log.Printf("[WARN] No such policy by name (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("policy_name", string(userInfo.PolicyName)); err != nil {
		return err
	}

	return nil
}

func minioDeleteUserPolicyAttachment(d *schema.ResourceData, meta interface{}) error {
	minioAdmin := meta.(*S3MinioClient).S3Admin
	var userName = d.Get("user_name").(string)

	errIam := minioAdmin.SetPolicy("", userName, false)
	if errIam != nil {
		return NewResourceError("Unable to delete user policy", userName, errIam)
	}

	return nil
}

func minioImportUserPolicyAttachment(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <user-name>/<policy_name>", d.Id())
	}

	userName := idParts[0]
	policyName := idParts[1]

	err := d.Set("user_name", userName)
	if err != nil {
		return nil, NewResourceError("Unable to import user policy", userName, err)
	}
	err = d.Set("policy_name", policyName)
	if err != nil {
		return nil, NewResourceError("Unable to import user policy", userName, err)
	}
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", userName)))
	return []*schema.ResourceData{d}, nil
}
