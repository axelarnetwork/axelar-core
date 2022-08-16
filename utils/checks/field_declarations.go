package checks

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"

	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/maps"
	"github.com/axelarnetwork/utils/slices"
)

const excludeFieldsFlag = "excludeFields"
const excludeTypesFlag = "excludeTypes"

// FieldDeclarations returns a command that checks if structs have been correctly initialized
func FieldDeclarations() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "decls package-path...",
		Args: cobra.MinimumNArgs(1),
		RunE: run,
	}

	cmd.Flags().StringSliceP(excludeFieldsFlag, "f", []string{}, "exclusion pattern for fields")
	cmd.Flags().StringSliceP(excludeTypesFlag, "t", []string{}, "exclusion pattern for types")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	config := &packages.Config{
		Mode: packages.NeedTypesInfo |
			packages.NeedTypes |
			packages.NeedName |
			packages.NeedSyntax,
		Dir: funcs.Must(filepath.Abs("./")),
		Env: append(os.Environ(), "GO111MODULE=on"),
	}

	pkgs, err := packages.Load(config, args...)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if pkg.Errors != nil {
			writeErrs(pkg, cmd.OutOrStderr())
			return fmt.Errorf("error loading package %s", pkg.Name)
		}
	}

	depVis := &deprecatedVisitor{DeprecatedFields: map[token.Pos]struct{}{}}
	declVisitor := missingDeclVisitor{
		types:                  nil,
		fset:                   nil,
		deprecatedFields:       depVis.DeprecatedFields,
		out:                    cmd.OutOrStdout(),
		currentPackage:         "",
		typeExclusionPatterns:  funcs.Must(cmd.Flags().GetStringSlice(excludeTypesFlag)),
		fieldExclusionPatterns: funcs.Must(cmd.Flags().GetStringSlice(excludeFieldsFlag)),
		rootDirs:               slices.Map(pkgs, func(pkg *packages.Package) string { return pkg.PkgPath }),
	}

	packages.Visit(pkgs, findDeprecatedFields(depVis), checkFieldDeclarations(declVisitor))
	return nil
}

func findDeprecatedFields(depVis *deprecatedVisitor) func(p *packages.Package) bool {
	return func(p *packages.Package) bool {

		for _, file := range p.Syntax {
			ast.Walk(depVis, file)
		}

		return true
	}
}

func writeErrs(p *packages.Package, out io.Writer) {
	for _, err := range p.Errors {
		_, err := out.Write([]byte(err.Error()))
		if err != nil {
			panic(err)
		}
	}
}

func checkFieldDeclarations(visitor missingDeclVisitor) func(p *packages.Package) {
	return func(p *packages.Package) {
		if !slices.Any(visitor.rootDirs, func(dir string) bool { return strings.HasPrefix(p.PkgPath, dir) }) {
			return
		}

		visitor.types = p.TypesInfo
		visitor.fset = p.Fset

		for _, file := range p.Syntax {
			ast.Walk(visitor, file)
		}
	}
}

type deprecatedVisitor struct {
	DeprecatedFields map[token.Pos]struct{}
}

func (d *deprecatedVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.Field:
		var comments []*ast.Comment
		if node.Comment != nil {
			comments = append(comments, node.Comment.List...)
		}
		if node.Doc != nil {
			comments = append(comments, node.Doc.List...)
		}

		if slices.Any(comments, func(c *ast.Comment) bool { return strings.HasPrefix(c.Text, "// Deprecated") }) {
			d.DeprecatedFields[node.Pos()] = struct{}{}
		}
	}

	return d
}

type missingDeclVisitor struct {
	types                  *types.Info
	fset                   *token.FileSet
	deprecatedFields       map[token.Pos]struct{}
	out                    io.Writer
	currentPackage         string
	typeExclusionPatterns  []string
	fieldExclusionPatterns []string
	rootDirs               []string
}

func (v missingDeclVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	if strings.Contains(v.fset.File(node.Pos()).Name(), ".pb.") {
		return nil
	}

	position := v.fset.Position(node.Pos()).String()

	switch node := node.(type) {
	case *ast.File:
		v.currentPackage = node.Name.Name
	case *ast.CompositeLit:
		s, ok := v.convertToRelevantStruct(node, v.typeExclusionPatterns...)
		if !ok {
			break
		}

		declarations := getFieldDeclarations(node)

		// if fields are not explicitly labelled, Go ensures that all fields are filled, so ignore that case here
		if len(declarations) == 0 {
			break
		}

		requiredFields := getRequiredFieldNames(s, v.deprecatedFields, v.currentPackage, v.fieldExclusionPatterns...)
		missingViolations := checkMissingFields(declarations, requiredFields)

		writeViolations(v.out, position, missingViolations)
	}

	return v
}

func writeViolations(out io.Writer, position string, violations []violationMsg) {
	for _, msg := range violations {
		_, err := out.Write([]byte(fmt.Sprintln(violation{Pos: position, Msg: msg})))
		if err != nil {
			panic(err)
		}
	}
}

func checkMissingFields(declaredFields []string, requiredFields []string) []violationMsg {
	mappedFields := slices.ToMap(declaredFields, funcs.Identity[string])

	var violations []violationMsg
	for _, field := range requiredFields {
		if !maps.Has(mappedFields, field) {
			violations = append(violations, violationMsg(fmt.Sprintf("missing field %s", field)))
		}
	}
	return violations
}

func getFieldDeclarations(node *ast.CompositeLit) []string {
	var declarations []string
	for _, elt := range node.Elts {
		switch elt := elt.(type) {
		case *ast.KeyValueExpr:
			declarations = append(declarations, elt.Key.(*ast.Ident).Name)
		}
	}

	return declarations
}

func getRequiredFieldNames(s *types.Struct, deprecatedFields map[token.Pos]struct{}, currentPackage string, exclusionPatterns ...string) []string {
	requiredFields := slices.Filter(getAllFields(s), isRequired(deprecatedFields, currentPackage, exclusionPatterns...))
	return slices.Map(requiredFields, (*types.Var).Name)
}

func getAllFields(s *types.Struct) []*types.Var {
	var fields []*types.Var
	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		fields = append(fields, field)
	}
	return fields
}

func isRequired(deprecatedFields map[token.Pos]struct{}, currentPackage string, exclusionPatterns ...string) func(*types.Var) bool {
	return func(field *types.Var) bool {
		if !field.Exported() && field.Pkg().Name() != currentPackage {
			return false
		}

		for _, pattern := range exclusionPatterns {
			if strings.Contains(field.Name(), pattern) {
				return false
			}
		}

		if _, ok := deprecatedFields[field.Pos()]; ok {
			return false
		}
		return true
	}
}

func (v missingDeclVisitor) convertToRelevantStruct(node *ast.CompositeLit, exclusionPatterns ...string) (*types.Struct, bool) {
	// ignore empty structs (e.g. as a return in a function that can return an error)
	if len(node.Elts) == 0 {
		return nil, false
	}

	t := v.types.TypeOf(node)
	s, ok := toStruct(t)
	if !ok {
		return nil, false
	}

	for _, pattern := range exclusionPatterns {
		if strings.Contains(t.String(), pattern) {
			return nil, false
		}
	}

	return s, true
}

func toStruct(t types.Type) (*types.Struct, bool) {
	switch t := t.(type) {
	case *types.Struct:
		return t, true
	case *types.Named:
		return toStruct(t.Underlying())
	case *types.Pointer:
		return toStruct(t.Elem())
	default:
		return nil, false
	}
}

type violationMsg string

type violation struct {
	Pos string
	Msg violationMsg
}

func (v violation) String() string {
	return fmt.Sprintf("%s: %s", v.Pos, v.Msg)
}
