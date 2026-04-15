package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	var inPath string
	var outPath string

	flag.StringVar(&inPath, "in", "", "input OpenAPI YAML path")
	flag.StringVar(&outPath, "out", "", "output normalized OpenAPI YAML path")
	flag.Parse()

	if inPath == "" || outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: openapi-normalize -in <input> -out <output>")
		os.Exit(2)
	}

	input, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(input, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "parse yaml: %v\n", err)
		os.Exit(1)
	}

	normalizeNode(&doc)
	if len(doc.Content) > 0 {
		extractOperationSchemas(doc.Content[0])
	}

	output, err := yaml.Marshal(&doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal yaml: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, output, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func normalizeNode(node *yaml.Node) {
	if node == nil {
		return
	}

	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			normalizeNode(child)
		}
		return
	}

	if node.Kind == yaml.MappingNode {
		normalizeOpenAPIVersion(node)
		normalizeNullableType(node)
		addOperationID(node)

		for _, child := range node.Content {
			normalizeNode(child)
		}
		return
	}

	for _, child := range node.Content {
		normalizeNode(child)
	}
}

func extractOperationSchemas(root *yaml.Node) {
	if root == nil || root.Kind != yaml.MappingNode {
		return
	}

	pathsNode := mappingValue(root, "paths")
	if pathsNode == nil || pathsNode.Kind != yaml.MappingNode {
		return
	}

	componentsNode := ensureMapping(root, "components")
	schemasNode := ensureMapping(componentsNode, "schemas")

	for i := 0; i < len(pathsNode.Content); i += 2 {
		pathValue := pathsNode.Content[i+1]
		if pathValue.Kind != yaml.MappingNode {
			continue
		}

		for j := 0; j < len(pathValue.Content); j += 2 {
			operationNode := pathValue.Content[j+1]
			if operationNode.Kind != yaml.MappingNode {
				continue
			}

			operationIDNode := mappingValue(operationNode, "operationId")
			if operationIDNode == nil || operationIDNode.Value == "" {
				continue
			}

			operationName := exportedName(operationIDNode.Value)
			extractRequestBodySchema(operationNode, schemasNode, operationName+"Request")
			extractResponseSchemas(operationNode, schemasNode, operationName)
		}
	}
}

func extractRequestBodySchema(operationNode, schemasNode *yaml.Node, schemaName string) {
	requestBodyNode := mappingValue(operationNode, "requestBody")
	if requestBodyNode == nil {
		return
	}

	contentNode := mappingValue(requestBodyNode, "content")
	if contentNode == nil {
		return
	}

	applicationJSONNode := mappingValue(contentNode, "application/json")
	if applicationJSONNode == nil {
		return
	}

	schemaNode := mappingValue(applicationJSONNode, "schema")
	if schemaNode == nil || isRefNode(schemaNode) {
		return
	}

	ensureSchemaComponent(schemasNode, schemaName, schemaNode)
	replaceWithRef(schemaNode, schemaName)
}

func extractResponseSchemas(operationNode, schemasNode *yaml.Node, operationName string) {
	responsesNode := mappingValue(operationNode, "responses")
	if responsesNode == nil || responsesNode.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(responsesNode.Content); i += 2 {
		statusCode := responsesNode.Content[i].Value
		responseNode := responsesNode.Content[i+1]
		if responseNode.Kind != yaml.MappingNode {
			continue
		}

		contentNode := mappingValue(responseNode, "content")
		if contentNode == nil {
			continue
		}

		applicationJSONNode := mappingValue(contentNode, "application/json")
		if applicationJSONNode == nil {
			continue
		}

		schemaNode := mappingValue(applicationJSONNode, "schema")
		if schemaNode == nil || isRefNode(schemaNode) {
			continue
		}

		schemaName := operationName + exportedName(statusCode) + "Response"
		ensureSchemaComponent(schemasNode, schemaName, schemaNode)
		replaceWithRef(schemaNode, schemaName)
	}
}

func normalizeOpenAPIVersion(node *yaml.Node) {
	if value := mappingValue(node, "openapi"); value != nil && strings.HasPrefix(value.Value, "3.1.") {
		value.Value = "3.0.3"
	}
}

func normalizeNullableType(node *yaml.Node) {
	typeNode := mappingValue(node, "type")
	if typeNode == nil || typeNode.Kind != yaml.SequenceNode || len(typeNode.Content) != 2 {
		return
	}

	types := []string{typeNode.Content[0].Value, typeNode.Content[1].Value}
	if !slices.Contains(types, "null") {
		return
	}

	nonNull := ""
	for _, t := range types {
		if t != "null" {
			nonNull = t
			break
		}
	}
	if nonNull == "" {
		return
	}

	typeNode.Kind = yaml.ScalarNode
	typeNode.Tag = "!!str"
	typeNode.Value = nonNull
	typeNode.Content = nil

	if mappingValue(node, "nullable") == nil {
		appendMapping(node, scalarNode("nullable"), boolNode(true))
	}
}

func addOperationID(node *yaml.Node) {
	if node.Kind != yaml.MappingNode {
		return
	}

	pathsNode := mappingValue(node, "paths")
	if pathsNode == nil || pathsNode.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(pathsNode.Content); i += 2 {
		pathKey := pathsNode.Content[i]
		pathValue := pathsNode.Content[i+1]
		if pathValue.Kind != yaml.MappingNode {
			continue
		}

		for j := 0; j < len(pathValue.Content); j += 2 {
			methodKey := pathValue.Content[j]
			methodValue := pathValue.Content[j+1]
			if methodValue.Kind != yaml.MappingNode || mappingValue(methodValue, "operationId") != nil {
				continue
			}

			operationID := buildOperationID(methodKey.Value, pathKey.Value)
			appendMapping(methodValue, scalarNode("operationId"), scalarNode(operationID))
		}
	}
}

func buildOperationID(method, path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	parts := []string{strings.ToLower(method)}

	for _, segment := range segments {
		switch {
		case segment == "":
			continue
		case strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}"):
			parts = append(parts, "by", toPascalCase(strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")))
		default:
			parts = append(parts, toPascalCase(segment))
		}
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return parts[0] + strings.Join(parts[1:], "")
}

func toPascalCase(raw string) string {
	raw = strings.TrimSpace(raw)
	replacer := strings.NewReplacer("-", " ", "_", " ", ".", " ")
	words := strings.Fields(replacer.Replace(raw))

	var builder strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}

		builder.WriteString(strings.ToUpper(word[:1]))
		if len(word) > 1 {
			builder.WriteString(word[1:])
		}
	}

	return builder.String()
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

func appendMapping(node *yaml.Node, key *yaml.Node, value *yaml.Node) {
	node.Content = append(node.Content, key, value)
}

func ensureMapping(node *yaml.Node, key string) *yaml.Node {
	value := mappingValue(node, key)
	if value != nil {
		return value
	}

	value = &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	appendMapping(node, scalarNode(key), value)
	return value
}

func ensureSchemaComponent(schemasNode *yaml.Node, schemaName string, schemaNode *yaml.Node) {
	if mappingValue(schemasNode, schemaName) != nil {
		return
	}

	appendMapping(schemasNode, scalarNode(schemaName), cloneNode(schemaNode))
}

func replaceWithRef(schemaNode *yaml.Node, schemaName string) {
	schemaNode.Kind = yaml.MappingNode
	schemaNode.Tag = "!!map"
	schemaNode.Value = ""
	schemaNode.Content = []*yaml.Node{
		scalarNode("$ref"),
		scalarNode("#/components/schemas/" + schemaName),
	}
}

func isRefNode(node *yaml.Node) bool {
	return node.Kind == yaml.MappingNode && mappingValue(node, "$ref") != nil
}

func scalarNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

func boolNode(value bool) *yaml.Node {
	nodeValue := "false"
	if value {
		nodeValue = "true"
	}

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!bool",
		Value: nodeValue,
	}
}

func cloneNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	cloned := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		Alias:       node.Alias,
		Content:     make([]*yaml.Node, 0, len(node.Content)),
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,
		Column:      node.Column,
	}

	for _, child := range node.Content {
		cloned.Content = append(cloned.Content, cloneNode(child))
	}

	return cloned
}

func exportedName(value string) string {
	if value == "" {
		return value
	}

	if len(value) == 1 {
		return strings.ToUpper(value)
	}

	return strings.ToUpper(value[:1]) + value[1:]
}
