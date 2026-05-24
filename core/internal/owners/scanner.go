package owners

import (
	"pcc-toolkit/core/internal/pcc"
)

type OwnerEntry struct {
	ConversationName string `json:"conversation"`
	OwnerTag         string `json:"owner"`
	File             string `json:"file"`
}

type OwnerOutput struct {
	File   string       `json:"file"`
	Owners []OwnerEntry `json:"owners"`
}

var startConversationClasses = map[string]bool{
	"BioSeqAct_StartConversation": true,
	"SFXSeqAct_StartConversation": true,
	"SFXSeqAct_StartAmbientConv":  true,
}

func ScanOwners(rawData []byte, summary *pcc.FileSummary, path string) *OwnerOutput {
	output := &OwnerOutput{File: path}

	for _, exp := range summary.Exports {
		if !startConversationClasses[exp.ClassName] {
			continue
		}

		props, _ := pcc.ParsePropertyCollection(rawData, summary.Names, exp.SerialOffset, exp.SerialSize)
		if props == nil {
			continue
		}

		convName := resolveObjectRefName(props, "Conv", rawData, summary)
		if convName == "" {
			continue
		}

		ownerTag := extractOwnerFromVariableLinks(props, "VariableLinks", rawData, summary)
		if ownerTag == "" {
			ownerTag = "Not found"
		}

		output.Owners = append(output.Owners, OwnerEntry{
			ConversationName: convName,
			OwnerTag:         ownerTag,
			File:             path,
		})
	}

	return output
}

func resolveObjectRefName(props map[string]pcc.ParsedProperty, propName string, rawData []byte, summary *pcc.FileSummary) string {
	prop, ok := props[propName]
	if !ok || prop.PropType != "ObjectProperty" {
		return ""
	}
	ref, ok := prop.Value.(int)
	if !ok {
		return ""
	}
	return resolveObjectName(ref, summary)
}

func resolveObjectName(objRef int, summary *pcc.FileSummary) string {
	if objRef > 0 {
		idx := objRef - 1
		if idx >= 0 && idx < len(summary.Exports) {
			return summary.Exports[idx].ObjectName
		}
	} else if objRef < 0 {
		idx := (-objRef) - 1
		if idx >= 0 && idx < len(summary.Imports) {
			nameIdx := summary.Imports[idx].ObjectNameIndex
			if nameIdx >= 0 && nameIdx < len(summary.Names) {
				return summary.Names[nameIdx]
			}
		}
	}
	return ""
}

func resolveObjectExport(objRef int, summary *pcc.FileSummary) *pcc.Export {
	if objRef > 0 {
		idx := objRef - 1
		if idx >= 0 && idx < len(summary.Exports) {
			return &summary.Exports[idx]
		}
	}
	return nil
}

func extractOwnerFromVariableLinks(props map[string]pcc.ParsedProperty, propName string, rawData []byte, summary *pcc.FileSummary) string {
	prop, ok := props[propName]
	if !ok || prop.PropType != "ArrayProperty" {
		return ""
	}

	vl, ok := prop.Value.(map[string]interface{})
	if !ok {
		return ""
	}
	count, _ := vl["count"].(int)
	payloadOffset, _ := vl["payload_offset"].(int)
	payloadSize, _ := vl["payload_size"].(int)

	if count <= 0 || payloadOffset <= 0 || payloadSize <= 0 {
		return ""
	}

	items := pcc.ParseStructArrayItemsAsPropertyCollections(rawData, summary.Names, payloadOffset, payloadSize, count)
	for _, item := range items {
		descProp, ok := item["LinkDesc"]
		if !ok {
			continue
		}
		desc, _ := descProp.Value.(string)
		if desc != "Owner" {
			continue
		}

		linkedVarsProp, ok := item["LinkedVariables"]
		if !ok || linkedVarsProp.PropType != "ArrayProperty" {
			continue
		}
		lv, ok := linkedVarsProp.Value.(map[string]interface{})
		if !ok {
			continue
		}
		lvCount, _ := lv["count"].(int)
		lvPayload, _ := lv["payload_offset"].(int)
		if lvCount <= 0 || lvPayload <= 0 {
			continue
		}

		objRef := pcc.ReadRawI32(rawData, lvPayload)
		return resolveOwnerTag(objRef, rawData, summary)
	}
	return ""
}

func resolveOwnerTag(objRef int, rawData []byte, summary *pcc.FileSummary) string {
	exp := resolveObjectExport(objRef, summary)
	if exp == nil {
		return "Not found"
	}

	props, _ := pcc.ParsePropertyCollection(rawData, summary.Names, exp.SerialOffset, exp.SerialSize)
	if props == nil {
		return "Not found"
	}

	switch exp.ClassName {
	case "SeqVar_Object":
		return extractOwnerFromSeqVarObject(props, rawData, summary)
	case "BioSeqVar_ObjectFindByTag":
		return extractOwnerFromObjectFindByTag(props)
	default:
		return "Not found"
	}
}

func extractOwnerFromSeqVarObject(props map[string]pcc.ParsedProperty, rawData []byte, summary *pcc.FileSummary) string {
	objValueProp, ok := props["ObjValue"]
	if !ok || objValueProp.PropType != "ObjectProperty" {
		return "Not found"
	}
	actorRef, ok := objValueProp.Value.(int)
	if !ok {
		return "Not found"
	}

	actorExp := resolveObjectExport(actorRef, summary)
	if actorExp == nil {
		return "Not found"
	}

	actorProps, _ := pcc.ParsePropertyCollection(rawData, summary.Names, actorExp.SerialOffset, actorExp.SerialSize)
	if actorProps == nil {
		return "Not found"
	}

	tagProp, ok := actorProps["Tag"]
	if !ok || tagProp.PropType != "NameProperty" {
		return "Not found"
	}
	tag, ok := tagProp.Value.(string)
	if ok {
		return tag
	}
	return "Not found"
}

func extractOwnerFromObjectFindByTag(props map[string]pcc.ParsedProperty) string {
	if tagProp, ok := props["m_sObjectTagToFind"]; ok {
		if tag, ok := tagProp.Value.(string); ok {
			return tag
		}
	}
	return "Not found"
}
