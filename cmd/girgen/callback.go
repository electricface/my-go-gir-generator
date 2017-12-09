package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/electricface/my-go-gir-generator/gi"
)

func pCallback(s *SourceFile, callback *gi.CallbackInfo) {
	name := callback.Name()
	defer func() {
		if err := recover(); err != nil {
			log.Println("pCallback", name)
			panic(err)
		}
	}()

	if callback.Parameters.InstanceParameter != nil {
		panic("assert failed callback.Parameters.InstanceParameter == nil")
	}
	var paramTpls []ParamTemplate
	var args []string
	for idx, param := range callback.Parameters.Parameters {
		tpl := newParamTemplate(param)
		if tpl == nil {
			panic("newParamTemplate failed for param " + param.Name)
		}

		paramTpls = append(paramTpls, tpl)

		if idx != param.ClosureIndex {
			args = append(args, getVarTypeForGo(tpl, false))
		} // else: param is user_data
	}

	argsJoined := strings.Join(args, ", ")
	returnType := ""
	// TODO handle return type
	s.GoBody.Pn("type %s func (%s) %s", name, argsJoined, returnType)

	pCallbackWrapper(s, callback)
}

func getClosureParamName(callback *gi.CallbackInfo) string {
	for idx, param := range callback.Parameters.Parameters {
		if idx == param.ClosureIndex {
			return param.Name
		}
	}
	panic("failed to get closure param name")
}

// static void AsyncReadyCallbackWrapper(GObject *source_object, GAsyncResult *res, gpointer user_data);
func pCallbackWrapper(s *SourceFile, callback *gi.CallbackInfo) {
	name := callback.Name()
	returnType := "void"
	var args []string

	for _, param := range callback.Parameters.Parameters {
		args = append(args, param.Type.CType+" "+param.Name)
	}

	argsJoined := strings.Join(args, ", ")

	funcHeader := fmt.Sprintf("static %s %sWrapper(%s)", returnType, name, argsJoined)
	s.CHeader.Pn("%s;", funcHeader)
	s.CBody.Pn("%s {", funcHeader)

	s.CBody.Pn("    GClosure* closure = %s;", getClosureParamName(callback))

	paramCount := len(callback.Parameters.Parameters) - 1
	s.CBody.Pn("    GValue params[%d];", paramCount)

	s.AddCInclude("<strings.h>")
	s.CBody.Pn("    bzero(params, %d*sizeof(GValue));", paramCount)

	var count int
	for idx, param := range callback.Parameters.Parameters {
		if idx != param.ClosureIndex {
			gtype, setter := getGValueTypeAndSetter(param.Type)

			s.CBody.Pn("    g_value_init(&params[%d], %s);", count, gtype)
			s.CBody.Pn("    g_value_set_%s(&params[%d], %s);", setter, count, param.Name)
			count++
		} // else: param is user_data
	}

	s.CBody.Pn("    g_closure_invoke(closure, NULL, %d, params, NULL);", paramCount)

	if _, ok := asyncCallbackMap[name]; ok {
		// is async callback
		s.CBody.Pn("    g_closure_unref(closure);")
	}

	s.CBody.Pn("}")
}

func getGValueTypeAndSetter(type0 *gi.Type) (string, string) {
	typeName := type0.Name
	typeDef, _ := repo.GetType(typeName)
	if typeDef != nil {
		//_, isEnum := typeDef.(*gi.EnumInfo)
		//_, isStruct := typeDef.(*gi.StructInfo)
		_, isObject := typeDef.(*gi.ObjectInfo)
		_, isInterface := typeDef.(*gi.InterfaceInfo)
		//_, isAlias := typeDef.(*gi.AliasInfo)
		//_, isCallback := typeDef.(*gi.CallbackInfo)
		if isObject {
			return "G_TYPE_OBJECT", "object"
		}

		if isInterface {
			ifc := typeDef.(*gi.InterfaceInfo)
			return ifc.GlibGetType + "()", "object"
		}
	}

	cType, err := gi.ParseCType(type0.CType)
	if err != nil {
		panic(err)
	}

	typeForC := cType.CgoNotation()
	key := typeForC + "," + typeName
	arr, ok := gValueTypeAndSetterMap[key]
	if !ok {
		panic("fail to get gvalue type and setter for " + key)
	}
	return arr[0], arr[1]
}

var gValueTypeAndSetterMap = map[string][2]string{
	"C.goffset,gint64": {
		"G_TYPE_INT64", "int64",
	},
	"C.gboolean,gboolean": {
		"G_TYPE_BOOLEAN", "boolean",
	},
	"C.gdouble,gdouble": {
		"G_TYPE_DOUBLE", "double",
	},
	"C.gfloat,gfloat": {
		"G_TYPE_FLOAT", "float",
	},
	"C.gint,gint": {
		"G_TYPE_INT", "int",
	},
	"C.glong,glong": {
		"G_TYPE_LONG", "long",
	},
	"C.gpointer,gpointer": {
		"G_TYPE_POINTER", "pointer",
	},
	"C.gint8,int8": {
		"G_TYPE_CHAR", "schar",
	},
	"*C.gchar,utf8": {
		"G_TYPE_STRING", "string",
	},
	"*C.char,utf8": {
		"G_TYPE_STRING", "string",
	},
	"C.guchar,guchar": {
		"G_TYPE_UCHAR", "uchar",
	},
	"C.guint,guint": {
		"G_TYPE_UINT", "uint",
	},
	"C.guint64,guint64": {
		"G_TYPE_UINT64", "uint64",
	},
	"C.gulong,gulong": {
		"G_TYPE_ULONG", "ulong",
	},
	"*C.GVariant,GLib.Variant": {
		"G_TYPE_VARIANT", "variant",
	},
}
