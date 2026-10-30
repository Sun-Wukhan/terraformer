// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package terraformoutput

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"

	"github.com/hashicorp/terraform/terraform"
)

func OutputHclFiles(resources []terraformutils.Resource, provider terraformutils.ProviderGenerator, path string, serviceName string, isCompact bool, output string, sort bool) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	providerConfig := map[string]interface{}{
		"version": providerwrapper.GetProviderVersion(provider.GetName()),
	}

	if providerWithSource, ok := provider.(terraformutils.ProviderWithSource); ok {
		providerConfig["source"] = providerWithSource.GetSource()
	} else if source := providerwrapper.GetProviderSource(provider.GetName()); source != "" {
		providerConfig["source"] = source
	}

	// create provider file
	providerData := provider.GetProviderData()
	providerData["terraform"] = map[string]interface{}{
		"required_providers": []map[string]interface{}{{
			provider.GetName(): providerConfig,
		}},
	}

	var providerDataFile []byte
	var err error
	if output == "hcl" {
		// The generic JSON->HCL1 conversion renders nested provider config
		// maps as extra block labels (e.g. `azurerm "features" {}`), which
		// is invalid. Render provider.tf with a dedicated modern writer.
		providerDataFile = printProviderTF(providerData)
	} else {
		providerDataFile, err = terraformutils.Print(providerData, map[string]struct{}{}, output, sort)
		if err != nil {
			return err
		}
	}
	PrintFile(path+"/provider."+GetFileExtension(output), providerDataFile)

	// create outputs files
	outputs := map[string]interface{}{}
	outputsByResource := map[string]map[string]interface{}{}

	for i, r := range resources {
		outputState := map[string]*terraform.OutputState{}
		outputsByResource[r.InstanceInfo.Type+"_"+r.ResourceName+"_"+r.GetIDKey()] = map[string]interface{}{
			"value": "${" + r.InstanceInfo.Type + "." + r.ResourceName + "." + r.GetIDKey() + "}",
		}
		outputState[r.InstanceInfo.Type+"_"+r.ResourceName+"_"+r.GetIDKey()] = &terraform.OutputState{
			Type:  "string",
			Value: r.InstanceState.Attributes[r.GetIDKey()],
		}
		for _, v := range provider.GetResourceConnections() {
			for k, ids := range v {
				if (serviceName != "" && k == serviceName) || (serviceName == "" && k == r.ServiceName()) {
					// ids is a flat list of (referencing attribute, referenced
					// attribute) pairs; every referenced attribute needs an
					// output, not just the first.
					for i := 1; i < len(ids); i += 2 {
						attr := ids[i]
						if _, exist := r.InstanceState.Attributes[attr]; !exist {
							continue
						}
						key := attr
						if attr == "self_link" || attr == "id" {
							key = r.GetIDKey()
						}
						linkKey := r.InstanceInfo.Type + "_" + r.ResourceName + "_" + key
						outputsByResource[linkKey] = map[string]interface{}{
							"value": "${" + r.InstanceInfo.Type + "." + r.ResourceName + "." + key + "}",
						}
						outputState[linkKey] = &terraform.OutputState{
							Type:  "string",
							Value: r.InstanceState.Attributes[attr],
						}
					}
				}
			}
		}
		resources[i].Outputs = outputState
	}
	if len(outputsByResource) > 0 {
		outputs["output"] = outputsByResource
		outputsFile, err := terraformutils.Print(outputs, map[string]struct{}{}, output, sort)
		if err != nil {
			return err
		}
		PrintFile(path+"/outputs."+GetFileExtension(output), outputsFile)
	}

	// group by resource by type
	typeOfServices := map[string][]terraformutils.Resource{}
	for _, r := range resources {
		typeOfServices[r.InstanceInfo.Type] = append(typeOfServices[r.InstanceInfo.Type], r)
	}
	if isCompact {
		err := printFile(resources, "resources", path, output, sort)
		if err != nil {
			return err
		}
	} else {
		for k, v := range typeOfServices {
			fileName := strings.ReplaceAll(k, strings.Split(k, "_")[0]+"_", "")
			err := printFile(v, fileName, path, output, sort)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printFile(v []terraformutils.Resource, fileName, path, output string, sort bool) error {
	for _, res := range v {
		if res.DataFiles == nil {
			continue
		}
		for fileName, content := range res.DataFiles {
			if err := os.MkdirAll(path+"/data/", os.ModePerm); err != nil {
				return err
			}
			err := os.WriteFile(path+"/data/"+fileName, content, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	tfFile, err := terraformutils.HclPrintResource(v, map[string]interface{}{}, output, sort)
	if err != nil {
		return err
	}
	err = os.WriteFile(path+"/"+fileName+"."+GetFileExtension(output), tfFile, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func PrintFile(path string, data []byte) {
	err := os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetFileExtension(outputFormat string) string {
	if outputFormat == "json" {
		return "tf.json"
	}
	return "tf"
}

// printProviderTF renders provider.tf in modern HCL: a terraform block with
// required_providers, followed by one provider block per provider with nested
// maps rendered as nested blocks (never as extra labels).
func printProviderTF(providerData map[string]interface{}) []byte {
	var b strings.Builder
	if tf, ok := providerData["terraform"].(map[string]interface{}); ok {
		b.WriteString("terraform {\n")
		if rp, ok := tf["required_providers"].([]map[string]interface{}); ok {
			// Merge the single-entry maps into one name->config map.
			merged := map[string]map[string]interface{}{}
			for _, m := range rp {
				for name, cfg := range m {
					if c, ok := cfg.(map[string]interface{}); ok {
						merged[name] = c
					}
				}
			}
			names := make([]string, 0, len(merged))
			for name := range merged {
				names = append(names, name)
			}
			sort.Strings(names)
			b.WriteString("  required_providers {\n")
			for _, name := range names {
				cfg := merged[name]
				b.WriteString(fmt.Sprintf("    %s = {\n", name))
				if src, ok := cfg["source"]; ok {
					b.WriteString(fmt.Sprintf("      source  = %s\n", hclLiteral(src)))
				}
				if ver, ok := cfg["version"]; ok {
					b.WriteString(fmt.Sprintf("      version = %s\n", hclLiteral(ver)))
				}
				b.WriteString("    }\n")
			}
			b.WriteString("  }\n")
		}
		b.WriteString("}\n")
	}
	if prov, ok := providerData["provider"].(map[string]interface{}); ok {
		names := make([]string, 0, len(prov))
		for name := range prov {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("provider %q {\n", name))
			if cfg, ok := prov[name].(map[string]interface{}); ok {
				// Version constraints belong in required_providers; inside a
				// provider block they have been deprecated since 0.13.
				body := make(map[string]interface{}, len(cfg))
				for k, v := range cfg {
					if k != "version" {
						body[k] = v
					}
				}
				writeHCLBlockBody(&b, body, 1)
			}
			b.WriteString("}\n")
		}
	}
	return []byte(b.String())
}

// writeHCLBlockBody renders a config map as HCL attributes and nested blocks.
func writeHCLBlockBody(b *strings.Builder, m map[string]interface{}, indent int) {
	pad := strings.Repeat("  ", indent)
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		switch v := m[name].(type) {
		case map[string]interface{}:
			if len(v) == 0 {
				b.WriteString(fmt.Sprintf("%s%s {}\n", pad, name))
			} else {
				b.WriteString(fmt.Sprintf("%s%s {\n", pad, name))
				writeHCLBlockBody(b, v, indent+1)
				b.WriteString(fmt.Sprintf("%s}\n", pad))
			}
		default:
			b.WriteString(fmt.Sprintf("%s%s = %s\n", pad, name, hclLiteral(v)))
		}
	}
}

// hclLiteral renders a scalar config value as an HCL literal.
func hclLiteral(v interface{}) string {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, hclLiteral(e))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", t)
	}
}
