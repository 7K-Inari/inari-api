package observability

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"go.yaml.in/yaml/v3"
)

func decodeDocument(data []byte, out any, label string) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return field(label, "document is empty")
	}
	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		return field(label, "invalid YAML: %v", err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return field(label, "expected one mapping document")
	}
	if err := rejectDuplicateKeys(root.Content[0], label); err != nil {
		return err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return field(label, "multiple documents are not allowed")
		}
		return field(label, "invalid YAML: %v", err)
	}
	dec = yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		return field(label, "invalid or unknown fields: %v", err)
	}
	return nil
}

func rejectDuplicateKeys(node *yaml.Node, path string) error {
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if seen[key] {
				return field(path, "duplicate key %q", key)
			}
			seen[key] = true
			if err := rejectDuplicateKeys(node.Content[i+1], path+"."+key); err != nil {
				return err
			}
		}
	}
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		for i, child := range node.Content {
			if err := rejectDuplicateKeys(child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func DecodeCapability(data []byte) (*Capability, error) {
	var capability Capability
	if err := decodeDocument(data, &capability, "capability"); err != nil {
		return nil, err
	}
	if err := capability.Validate(); err != nil {
		return nil, err
	}
	return &capability, nil
}

func DecodeVerificationResult(data []byte) (*VerificationResult, error) {
	var result VerificationResult
	if err := decodeDocument(data, &result, "verification"); err != nil {
		return nil, err
	}
	if err := result.ValidateAt(time.Now()); err != nil {
		return nil, err
	}
	return &result, nil
}
