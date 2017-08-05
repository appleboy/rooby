package vm

import (
	"github.com/st0012/metago"
	"plugin"
	"reflect"
)

func (vm *VM) initPluginObject(fn string, p *plugin.Plugin) *PluginObject {
	return &PluginObject{fn: fn, plugin: p, baseObj: &baseObj{class: vm.topLevelClass(pluginClass)}}
}

func (vm *VM) initPluginClass() *RClass {
	pc := vm.initializeClass(pluginClass, false)
	pc.setBuiltInMethods(builtinPluginClassMethods(), true)
	pc.setBuiltInMethods(builtinPluginInstanceMethods(), false)
	vm.objectClass.setClassConstant(pc)
	return pc
}

// PluginObject is a special type that contains a Go's plugin
type PluginObject struct {
	*baseObj
	fn     string
	plugin *plugin.Plugin
}

// Polymorphic helper functions -----------------------------------------
func (p *PluginObject) toString() string {
	return "<Plugin: " + p.fn + ">"
}

func (p *PluginObject) toJSON() string {
	return p.toString()
}

func builtinPluginClassMethods() []*BuiltInMethodObject {
	return []*BuiltInMethodObject{}
}

func builtinPluginInstanceMethods() []*BuiltInMethodObject {
	return []*BuiltInMethodObject{
		{
			Name: "send",
			Fn: func(receiver Object) builtinMethodBody {
				return func(t *thread, args []Object, blockFrame *callFrame) Object {
					s, ok := args[0].(*StringObject)

					if !ok {
						return t.vm.initErrorObject(TypeError, WrongArgumentTypeFormat, stringClass, args[0].Class().Name)
					}

					funcName := s.value
					r := receiver.(*PluginObject)
					p := r.plugin
					f, err := p.Lookup(funcName)

					if err != nil {
						return t.vm.initErrorObject(InternalError, err.Error())
					}

					funcArgs, err := convertToGoFuncArgs(args[1:])

					if err != nil {
						t.vm.initErrorObject(TypeError, err.Error())
					}

					result := reflect.ValueOf(reflect.ValueOf(f).Call(metago.WrapArguments(funcArgs...))).Interface()

					return t.vm.initObjectFromGoType(metago.UnwrapReflectValues(result))
				}
			},
		},
	}
}
