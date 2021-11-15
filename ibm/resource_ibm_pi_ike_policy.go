// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	st "github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/power-go-client/power/models"
)

const (
	PIPolicyId = "policy_id"
)

func resourceIBMPIIKEPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMPIIKEPolicyCreate,
		ReadContext:   resourceIBMPIIKEPolicyRead,
		UpdateContext: resourceIBMPIIKEPolicyUpdate,
		DeleteContext: resourceIBMPIIKEPolicyDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Required Attributes
			helpers.PICloudInstanceId: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "PI cloud instance ID",
			},
			helpers.PIVPNPolicyName: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the IKE Policy",
			},
			helpers.PIVPNPolicyDhGroup: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateAllowedIntValue([]int{1, 2, 5, 14, 19, 20, 24}),
				Description:  "DH group of the IKE Policy",
			},
			helpers.PIVPNPolicyEncryption: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAllowedStringValue([]string{"3des-cbc", "aes-128-cbc", "aes-128-gcm", "aes-192-cbc", "aes-256-cbc", "aes-256-gcm", "des-cbc"}),
				Description:  "Encryption of the IKE Policy",
			},
			helpers.PIVPNPolicyKeyLifetime: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateAllowedRangeInt(180, 86400),
				Description:  "Policy key lifetime",
			},
			helpers.PIVPNPolicyVersion: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validateAllowedRangeInt(1, 2),
				Description:  "Version of the IKE Policy",
			},
			helpers.PIVPNPolicyPresharedKey: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Preshared key used in this IKE Policy (length of preshared key must be even)",
			},

			// Optional Attributes
			helpers.PIVPNPolicyAuthentication: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validateAllowedStringValue([]string{"none", "sha-256", "sha-384", "sha1"}),
				Description:  "Authentication for the IKE Policy",
			},

			//Computed Attributes
			PIPolicyId: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IKE Policy ID",
			},
		},
	}
}

func resourceIBMPIIKEPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(helpers.PICloudInstanceId).(string)
	name := d.Get(helpers.PIVPNPolicyName).(string)
	dhGroup := int64(d.Get(helpers.PIVPNPolicyDhGroup).(int))
	encryption := d.Get(helpers.PIVPNPolicyEncryption).(string)
	keyLifetime := int64(d.Get(helpers.PIVPNPolicyKeyLifetime).(int))
	presharedKey := d.Get(helpers.PIVPNPolicyPresharedKey).(string)
	version := int64(d.Get(helpers.PIVPNPolicyVersion).(int))

	body := &models.IKEPolicyCreate{
		DhGroup:      &dhGroup,
		Encryption:   &encryption,
		KeyLifetime:  models.KeyLifetime(keyLifetime),
		Name:         &name,
		PresharedKey: &presharedKey,
		Version:      &version,
	}

	if v, ok := d.GetOk(helpers.PIVPNPolicyAuthentication); ok {
		body.Authentication = models.IKEPolicyAuthentication(v.(string))
	}

	client := st.NewIBMPIVpnPolicyClient(sess, cloudInstanceID)
	ikePolicy, err := client.CreateIKEPolicyWithContext(ctx, body, cloudInstanceID)
	if err != nil {
		log.Printf("[DEBUG] create ike policy failed %v", err)
		return diag.Errorf(errors.CreateVPNPolicyOperationFailed, cloudInstanceID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", cloudInstanceID, *ikePolicy.ID))

	return resourceIBMPIIKEPolicyRead(ctx, d, meta)
}

func resourceIBMPIIKEPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	policyID := parts[1]

	client := st.NewIBMPIVpnPolicyClient(sess, cloudInstanceID)
	body := &models.IKEPolicyUpdate{}

	if d.HasChange(helpers.PIVPNPolicyName) {
		name := d.Get(helpers.PIVPNPolicyName).(string)
		body.Name = name
	}
	if d.HasChange(helpers.PIVPNPolicyDhGroup) {
		dhGroup := int64(d.Get(helpers.PIVPNPolicyDhGroup).(int))
		body.DhGroup = dhGroup
	}
	if d.HasChange(helpers.PIVPNPolicyEncryption) {
		encryption := d.Get(helpers.PIVPNPolicyEncryption).(string)
		body.Encryption = encryption
	}
	if d.HasChange(helpers.PIVPNPolicyKeyLifetime) {
		keyLifetime := int64(d.Get(helpers.PIVPNPolicyKeyLifetime).(int))
		body.KeyLifetime = models.KeyLifetime(keyLifetime)
	}
	if d.HasChange(helpers.PIVPNPolicyPresharedKey) {
		presharedKey := d.Get(helpers.PIVPNPolicyPresharedKey).(string)
		body.PresharedKey = presharedKey
	}
	if d.HasChange(helpers.PIVPNPolicyVersion) {
		version := int64(d.Get(helpers.PIVPNPolicyVersion).(int))
		body.Version = version
	}
	if d.HasChange(helpers.PIVPNPolicyAuthentication) {
		authentication := d.Get(helpers.PIVPNPolicyAuthentication).(string)
		body.Authentication = models.IKEPolicyAuthentication(authentication)
	}

	_, err = client.UpdateIKEPolicyWithContext(ctx, body, policyID, cloudInstanceID)
	if err != nil {
		return diag.Errorf(errors.UpdateVPNPolicyOperationFailed, policyID, err)
	}

	return resourceIBMPIIKEPolicyRead(ctx, d, meta)
}

func resourceIBMPIIKEPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	policyID := parts[1]

	client := st.NewIBMPIVpnPolicyClient(sess, cloudInstanceID)
	ikePolicy, err := client.GetIKEPolicyWithContext(ctx, policyID, cloudInstanceID)
	if err != nil {
		// FIXME: Uncomment when 404 error is available
		// switch err.(type) {
		// case *p_cloud_v_p_n_policies.PcloudIkepoliciesGetNotFound:
		// 	log.Printf("[DEBUG] VPN policy does not exist %v", err)
		// 	d.SetId("")
		// 	return nil
		// }
		log.Printf("[DEBUG] get VPN policy failed %v", err)
		return diag.Errorf(errors.GetCloudConnectionOperationFailed, policyID, err)
	}

	d.Set(PIPolicyId, ikePolicy.ID)
	d.Set(helpers.PIVPNPolicyName, ikePolicy.Name)
	d.Set(helpers.PIVPNPolicyDhGroup, ikePolicy.DhGroup)
	d.Set(helpers.PIVPNPolicyEncryption, ikePolicy.Encryption)
	d.Set(helpers.PIVPNPolicyKeyLifetime, ikePolicy.KeyLifetime)
	d.Set(helpers.PIVPNPolicyVersion, ikePolicy.Version)
	d.Set(helpers.PIVPNPolicyAuthentication, ikePolicy.Authentication)

	return nil
}

func resourceIBMPIIKEPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	parts, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := parts[0]
	policyID := parts[1]

	client := st.NewIBMPIVpnPolicyClient(sess, cloudInstanceID)

	err = client.DeleteIKEPolicyWithContext(ctx, policyID, cloudInstanceID)
	if err != nil {
		// FIXME: Uncomment when 404 error is available
		// switch err.(type) {
		// case *p_cloud_v_p_n_policies.PcloudIkepoliciesDeleteNotFound:
		// 	log.Printf("[DEBUG] VPN policy does not exist %v", err)
		// 	d.SetId("")
		// 	return nil
		// }
		log.Printf("[DEBUG] delete VPN policy failed %v", err)
		return diag.Errorf(errors.DeleteVPNPolicyOperationFailed, policyID, err)
	}

	d.SetId("")
	return nil
}
