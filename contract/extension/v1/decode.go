package extensionv1

import (
	"bytes"
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

// DecodeSpec extracts spec.contracts from a full extension document or a spec
// object. Unknown additive fields are ignored by YAML decoding, while
// duplicate keys and multiple documents are rejected to avoid ambiguity.
// A missing contracts declaration returns (nil, nil).
func DecodeSpec(data []byte) (*Contracts, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, NewContractError(CodeContractMalformed, "file", "document is empty")
	}
	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		return nil, NewContractError(CodeContractMalformed, "file", "invalid YAML/JSON: %v", err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, NewContractError(CodeContractMalformed, "file", "expected one mapping document")
	}
	if err := rejectDuplicateKeys(root.Content[0], "file"); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, NewContractError(CodeContractMalformed, "file", "multiple documents are not allowed")
		}
		return nil, NewContractError(CodeContractMalformed, "file", "invalid YAML/JSON: %v", err)
	}

	node := root.Content[0]
	if spec := mappingValue(node, "spec"); spec != nil {
		node = spec
	}
	if node.Kind != yaml.MappingNode {
		return nil, NewContractError(CodeContractMalformed, "spec", "expected a mapping")
	}
	contractsNode := mappingValue(node, "contracts")
	if contractsNode == nil {
		return nil, nil
	}
	if contractsNode.Kind != yaml.MappingNode {
		return nil, NewContractError(CodeContractMalformed, "spec.contracts", "expected a mapping")
	}
	var contracts Contracts
	if err := contractsNode.Decode(&contracts); err != nil {
		return nil, NewContractError(CodeContractMalformed, "spec.contracts", "invalid declaration: %v", err)
	}
	if err := contracts.Validate(); err != nil {
		return nil, err
	}
	return &contracts, nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func rejectDuplicateKeys(node *yaml.Node, path string) error {
	if node.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if seen[key] {
				return NewContractError(CodeContractMalformed, path, "duplicate key %q", key)
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
