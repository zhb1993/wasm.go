package aot

type moduleCompiler struct {
	printer
	moduleInfo
}

func (c *moduleCompiler) compile() {
	c.genModule()
	c.genNew()
	c.println("")
	c.genExternalFuncs()
	c.genInternalFuncs()
	c.genUtils()
}

func (c *moduleCompiler) genModule() {
	c.print(`// Code generated by wasm.go. DO NOT EDIT.

package main

import (
	"math"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
)

type aotModule struct {
	importedFuncs []instance.Function
	table         instance.Table
	memory        instance.Memory
	globals       []instance.Global
}
`)
}

func (c *moduleCompiler) genNew() {
	funcCount := len(c.importedFuncs)
	globalCount := len(c.importedGlobals) + len(c.module.GlobalSec)
	c.printf(`
func Instantiate(iMap instance.Map) instance.Instance {
	m := &aotModule{
		importedFuncs: make([]instance.Function, %d),
		globals:       make([]instance.Global, %d),
	}
`, funcCount, globalCount)

	for i, imp := range c.importedFuncs {
		ft := c.module.TypeSec[imp.Desc.FuncType]
		c.printf(`	m.importedFuncs[%d] = iMap["%s"].Get("%s").(instance.Function) // %s%s`,
			i, imp.Module, imp.Name, ft.GetSignature(), "\n")
	}
	if len(c.importedTables) > 0 {
		c.printf(`	m.table = iMap["%s"].Get("%s").(instance.Table)%s`,
			c.importedTables[0].Module, c.importedTables[0].Name, "\n")
	} else {
		c.printf(`	m.table = interpreter.NewTable()%s`, "\n") // TODO
	}
	if len(c.importedMemories) > 0 {
		c.printf(`	m.memory = iMap["%s"].Get("%s").(instance.Memory)%s`,
			c.importedTables[0].Module, c.importedTables[0].Name, "\n")
	} else {
		c.printf(`	m.memory = interpreter.NewMemory()%s`, "\n") // TODO
	}
	for i, imp := range c.importedGlobals {
		c.printf(`	m.globals[%d] = iMap["%s"].Get("%s").(instance.Global)%s`,
			i, imp.Module, imp.Name, "\n")
	}
	for i, _ := range c.module.GlobalSec {
		c.printf(`	m.globals[%d] = interpreter.NewGlobal()%s`, // TODO
			len(c.importedGlobals)+i, "\n")
	}

	c.println("	return m\n}")
}

func (c *moduleCompiler) genExternalFuncs() {
	for i, imp := range c.importedFuncs {
		ft := c.module.TypeSec[imp.Desc.FuncType]

		fc := newFuncCompiler(c.moduleInfo)
		fc.printf("func (m *aotModule) f%d(", i)
		fc.genParams(len(ft.ParamTypes))
		fc.print(")")
		fc.genResults(len(ft.ResultTypes))
		fc.print(" {\n")
		fc.printf("	m.importedFuncs[%d]()", i)
		fc.println("}")
	}
}

func (c *moduleCompiler) genInternalFuncs() {
	importedFuncCount := len(c.importedFuncs)
	for i, ftIdx := range c.module.FuncSec {
		fc := newFuncCompiler(c.moduleInfo)
		fIdx := importedFuncCount + i
		ft := c.module.TypeSec[ftIdx]
		code := c.module.CodeSec[i]
		c.println(fc.compile(fIdx, ft, code))
	}
}

func (c *moduleCompiler) genUtils() {
	c.print(`// utils
func b2i(b bool) uint64 { if b { return 1 } else { return 0 } }
func f32(i uint64) float32 { return math.Float32frombits(uint32(i)) }
func u32(f float32) uint64 { return uint64(math.Float32bits(f)) }
func f64(i uint64) float64 { return math.Float64frombits(i) }
func u64(f float64) uint64 { return math.Float64bits(f) }
`)
}
