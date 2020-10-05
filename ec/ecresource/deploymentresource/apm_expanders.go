// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"reflect"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"

	"github.com/elastic/terraform-provider-ec/ec/internal/util"
)

var emptyApmConfig = &models.ApmConfiguration{
	SystemSettings: &models.ApmSystemSettings{},
}

// expandApmResources expands apm resources into their models.
func expandApmResources(apms []interface{}) ([]*models.ApmPayload, error) {
	if len(apms) == 0 {
		return nil, nil
	}

	result := make([]*models.ApmPayload, 0, len(apms))
	for _, raw := range apms {
		resResource, err := expandApmResource(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, resResource)
	}

	return result, nil
}

func expandApmResource(raw interface{}) (*models.ApmPayload, error) {
	var es = raw.(map[string]interface{})
	var res = models.ApmPayload{
		Plan: &models.ApmPlan{
			Apm: &models.ApmConfiguration{},
		},
		Settings: &models.ApmSettings{},
	}

	if esRefID, ok := es["elasticsearch_cluster_ref_id"]; ok {
		res.ElasticsearchClusterRefID = ec.String(esRefID.(string))
	}

	if refID, ok := es["ref_id"]; ok {
		res.RefID = ec.String(refID.(string))
	}

	if version, ok := es["version"]; ok {
		res.Plan.Apm.Version = version.(string)
	}

	if region, ok := es["region"]; ok {
		if r := region.(string); r != "" {
			res.Region = ec.String(r)
		}
	}

	if cfg, ok := es["config"]; ok {
		if c := expandApmConfig(cfg); c != nil {
			version := res.Plan.Apm.Version
			res.Plan.Apm = c
			res.Plan.Apm.Version = version
		}
	}

	if rawTopology, ok := es["topology"]; ok {
		topology, err := expandApmTopology(rawTopology)
		if err != nil {
			return nil, err
		}
		res.Plan.ClusterTopology = topology
	}

	return &res, nil
}

func expandApmTopology(raw interface{}) ([]*models.ApmTopologyElement, error) {
	var rawTopologies = raw.([]interface{})
	var res = make([]*models.ApmTopologyElement, 0, len(rawTopologies))
	for _, rawTop := range rawTopologies {
		var topology = rawTop.(map[string]interface{})

		size, err := util.ParseTopologySize(topology)
		if err != nil {
			return nil, err
		}

		var elem = models.ApmTopologyElement{
			Size: &size,
		}

		if id, ok := topology["instance_configuration_id"]; ok {
			elem.InstanceConfigurationID = id.(string)
		}

		if zones, ok := topology["zone_count"]; ok {
			elem.ZoneCount = int32(zones.(int))
		}

		if c, ok := topology["config"]; ok {
			elem.Apm = expandApmConfig(c)
		}

		res = append(res, &elem)
	}

	return res, nil
}

func expandApmConfig(raw interface{}) *models.ApmConfiguration {
	var res = &models.ApmConfiguration{
		SystemSettings: &models.ApmSystemSettings{},
	}
	for _, rawCfg := range raw.([]interface{}) {
		var cfg = rawCfg.(map[string]interface{})

		if debugEnabled, ok := cfg["debug_enabled"]; ok {
			res.SystemSettings.DebugEnabled = ec.Bool(debugEnabled.(bool))
		}

		if settings, ok := cfg["user_settings_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_override_json"]; ok && settings != nil {
			if s, ok := settings.(string); ok && s != "" {
				res.UserSettingsOverrideJSON = settings
			}
		}
		if settings, ok := cfg["user_settings_yaml"]; ok {
			res.UserSettingsYaml = settings.(string)
		}
		if settings, ok := cfg["user_settings_override_yaml"]; ok {
			res.UserSettingsOverrideYaml = settings.(string)
		}
	}

	if !reflect.DeepEqual(res, emptyApmConfig) {
		return res
	}

	return nil
}