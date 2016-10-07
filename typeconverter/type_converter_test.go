package typeconverter

import (
    //"fmt"
    "mygi"
    //"strings"
    //"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var emptyType = &mygi.Type{}
var emptyArray = &mygi.Array{}

//<parameter name="atomic" transfer-ownership="none">
  //<doc xml:space="preserve">a pointer to a #gint or #guint</doc>
  //<type name="gint" c:type="volatile gint*"/>
//</parameter>
func TestParamVolatilePointer(t *testing.T) {
    Convey("ParamVolatilePointer", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "gint",
            CType: "volatile gint*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*mygibase.Gint")
        So(cvt.CgoType, ShouldEqual, "*C.gint")
    })
}

//<parameter name="atomic" transfer-ownership="none">
  //<doc xml:space="preserve">a pointer to a #gint or #guint</doc>
  //<type name="gint" c:type="volatile const gint*"/>
//</parameter>
func TestParamVolatileConstPointer(t *testing.T) {
    Convey("ParamVolatileConstPointer", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "gint",
            CType: "volatile const gint*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*mygibase.Gint")
        So(cvt.CgoType, ShouldEqual, "*C.gint")
    })
}

//<parameter name="strv" transfer-ownership="none">
  //<doc xml:space="preserve">a %NULL-terminated array of strings</doc>
  //<type name="utf8" c:type="const gchar* const*"/>
//</parameter>
func TestParamConstPointer2(t *testing.T) {
    Convey("ParamConstPointer2", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "utf8",
            CType: "const gchar* const*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "string")
        So(cvt.CgoType, ShouldEqual, "**C.gchar")
    })
}

//<return-value transfer-ownership="full">
  //<array c:type="gchar**">
    //<type name="utf8"/>
  //</array>
//</return-value>
func TestReturnArrayString(t *testing.T) {
    Convey("ReturnArrayString",t,func(){
        cvt := NewTypeConverter(true, "", emptyType, &mygi.Array{
            CType: "char**",
            ElemType: mygi.ElemType{
                Name: "utf8",
            },
        })
        // gotype []string
        So(cvt.GoType, ShouldEqual, "[]string")
        So(cvt.CgoType, ShouldEqual, "**C.char")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "string")
        So(cvt.ElemCgoType, ShouldEqual, "*C.char")
    })
}

//<parameter name="people" transfer-ownership="none">
            //<doc xml:space="preserve">The people who belong to that section</doc>
            //<array c:type="gchar**">
              //<type name="utf8" c:type="gchar*"/>
            //</array>
          //</parameter>
// zero terminated true?
func TestParamArrayString(t *testing.T) {
    Convey("ParamArrayString", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "gchar**",
            ElemType: mygi.ElemType{
                Name: "utf8",
                CType:"gchar*",
            },
        })
        // gotype []string
        //fmt.Println("ParamArrayString", cvt)
        So(cvt.GoType, ShouldEqual, "[]string")
        So(cvt.CgoType, ShouldEqual, "**C.gchar")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "string")
        So(cvt.ElemCgoType, ShouldEqual, "*C.gchar")
    })
}


//<parameter name="argv"
                     //transfer-ownership="none"
                     //nullable="1"
                     //allow-none="1">
            //<doc xml:space="preserve">the argv from main(), or %NULL</doc>
            //<array length="0" zero-terminated="0" c:type="char**">
              //<type name="utf8" c:type="char*"/>
            //</array>
          //</parameter>
// zere terminated false
func TestParamArrayString1(t *testing.T) {
    Convey("ParamArrayString1", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "char**",
            ElemType: mygi.ElemType{
                Name: "utf8",
                CType:"char*",
            },
        })
        // gotype []string
        //fmt.Println("ParamArrayString1", cvt)
        So(cvt.GoType, ShouldEqual, "[]string")
        So(cvt.CgoType, ShouldEqual, "**C.char")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "string")
        So(cvt.ElemCgoType, ShouldEqual, "*C.char")
    })
}

//<parameter name="arguments"
           //direction="inout"
           //caller-allocates="0"
           //transfer-ownership="full">
  //<doc xml:space="preserve">array of command line arguments</doc>
  //<array c:type="gchar***">
    //<type name="utf8" c:type="gchar**"/>
  //</array>
//</parameter>
func TestParamArrayString2(t *testing.T) {
    Convey("ParamArrayString2", t, func(){
        cvt := NewTypeConverter(false, "inout", emptyType, &mygi.Array{
            CType: "gchar***",
            ElemType: mygi.ElemType{
                Name: "utf8",
                CType:"gchar**",
            },
        })
        // gotype []string
        //fmt.Println("ParamArrayString2", cvt)
        So(cvt.GoType, ShouldEqual, "[]string")
        // TODO handle direction inout
        So(cvt.CgoType, ShouldEqual, "***C.gchar")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "string")
    })
}

//<parameter name="entries" transfer-ownership="none">
//<doc xml:space="preserve">a pointer to
//the first item in an array of #GActionEntry structs</doc>
//<array length="1" zero-terminated="0" c:type="GActionEntry*">
  //<type name="ActionEntry"/>
//</array>
//</parameter>
func TestParamArrayStruct1(t *testing.T) {
    Convey("ParamArrayStruct1", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "GActionEntry*",
            ElemType: mygi.ElemType{
                Name: "ActionEntry",
                CType: "",
            },
        })
        // gotype []ActionEntry
        //fmt.Println("ParamArrayStruct1", cvt)
        So(cvt.GoType, ShouldEqual, "[]ActionEntry")
        So(cvt.CgoType, ShouldEqual, "*C.GActionEntry")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "ActionEntry")
        So(cvt.ElemCgoType, ShouldEqual, "C.GActionEntry")
    })
}

//<parameter name="files" transfer-ownership="none">
//<doc xml:space="preserve">an array of #GFiles to open</doc>
//<array length="1" zero-terminated="0" c:type="GFile**">
  //<type name="File" c:type="GFile*"/>
//</array>
//</parameter>
func TestParamArrayStruct2(t *testing.T) {
    Convey("ParamArrayStruct2", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "GFile**",
            ElemType: mygi.ElemType{
                Name: "File",
                CType: "GFile*",
            },
        })
        // gotype []*File
        //fmt.Println("ParamArrayStruct2", cvt)
        So(cvt.GoType, ShouldEqual, "[]*File")
        So(cvt.CgoType, ShouldEqual, "**C.GFile")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "*File")
        So(cvt.ElemCgoType, ShouldEqual, "*C.GFile")
    })
}

//<parameter name="entries" transfer-ownership="none">
            //<doc xml:space="preserve">a
          //%NULL-terminated list of #GOptionEntrys</doc>
            //<array c:type="GOptionEntry*">
              //<type name="GLib.OptionEntry"/>
            //</array>
          //</parameter>
func TestParamArrayStruct3(t *testing.T) {
    Convey("ParamArrayStruct3", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "GOptionEntry*",
            ElemType: mygi.ElemType{
                Name: "GLib.OptionEntry",
                CType: "",
            },
        })
        // gotype []glib.OptionEntry
        //fmt.Println("ParamArrayStruct3", cvt)
        So(cvt.GoType, ShouldEqual, "[]glib.OptionEntry")
        So(cvt.CgoType, ShouldEqual, "*C.GOptionEntry")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "glib.OptionEntry")
        So(cvt.ElemCgoType, ShouldEqual, "C.GOptionEntry")
    })
}

  //<parameter name="parameters" transfer-ownership="none">
    //<doc xml:space="preserve">the parameters to use to construct the object</doc>
    //<array length="1" zero-terminated="0" c:type="GParameter*">
      //<type name="GObject.Parameter" c:type="GParameter"/>
    //</array>
  //</parameter>
func TestParamArrayStruct4(t *testing.T) {
    Convey("ParamArrayStruct4",t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "GParameter*",
            ElemType: mygi.ElemType{
                Name: "GObject.Parameter",
                CType: "GParameter",
            },
        })
        // gotype []gobject.Parameter
        //fmt.Println("ParamArrayStruct4", cvt)
        So(cvt.GoType, ShouldEqual, "[]gobject.Parameter")
        So(cvt.CgoType, ShouldEqual, "*C.GParameter")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "gobject.Parameter")
        So(cvt.ElemCgoType, ShouldEqual, "C.GParameter")
    })
}

  //<parameter name="messages" transfer-ownership="none">
    //<doc xml:space="preserve">an array of #GInputMessage structs</doc>
    //<array length="1" zero-terminated="0" c:type="GInputMessage*">
      //<type name="InputMessage" c:type="GInputMessage"/>
    //</array>
  //</parameter>
func TestParamArrayStruct(t *testing.T) {
    Convey("TestParamArrayStruct", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "GInputMessage*",
            ElemType: mygi.ElemType{
                Name: "InputMessage",
                CType: "GInputMessage",
            },
        })
        // gotype []InputMessage
        //fmt.Println("ParamArrayStruct", cvt)
        So(cvt.GoType, ShouldEqual, "[]InputMessage")
        So(cvt.CgoType, ShouldEqual, "*C.GInputMessage")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "InputMessage")
        So(cvt.ElemCgoType, ShouldEqual, "C.GInputMessage")
    })
}
      //<parameter name="buffer" transfer-ownership="none">
        //<doc xml:space="preserve">a pointer to
//an allocated chunk of memory</doc>
        //<array length="2" zero-terminated="0" c:type="void*">
          //<type name="guint8"/>
        //</array>
      //</parameter>
func TestParamArrayGuint8(t *testing.T) {
    Convey("ParamArrayGuint8", t, func(){
        cvt := NewTypeConverter(false, "", emptyType, &mygi.Array{
            CType: "void*",
            ElemType: mygi.ElemType{
                Name: "guint8",
                CType: "",
            },
        })
        // gotype []uint8
        //fmt.Println("ParamArrayGuint8", cvt)
        So(cvt.GoType, ShouldEqual, "[]uint8")
        So(cvt.CgoType, ShouldEqual, "unsafe.Pointer")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "uint8")
        // TODO
        //So(cvt.ElemCgoType, ShouldEqual, "")
    })
}
    //<return-value transfer-ownership="none">
      //<doc xml:space="preserve">
     //read-only buffer</doc>
      //<array length="0" zero-terminated="0" c:type="void*">
        //<type name="guint8"/>
      //</array>
    //</return-value>
func TestReturnArrayGuint8_1(t *testing.T) {
    Convey("ReturnArrayGuint8_1",t,func(){
        cvt := NewTypeConverter(true, "", emptyType, &mygi.Array{
            CType: "void*",
            ElemType: mygi.ElemType{
                Name: "guint8",
                CType: "",
            },
        })
        // gotype []uint8
        //fmt.Println("ReturnArrayGuint8_1", cvt)
        So(cvt.GoType, ShouldEqual, "[]uint8")
        So(cvt.CgoType, ShouldEqual, "unsafe.Pointer")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "uint8")
    })
}
        //<return-value transfer-ownership="none">
          //<doc xml:space="preserve">An array of header fields
//terminated by %G_DBUS_MESSAGE_HEADER_FIELD_INVALID.  Each element
//is a #guchar. Free with g_free().</doc>
          //<array c:type="guchar*">
            //<type name="guint8" c:type="guchar"/>
          //</array>
        //</return-value>
func TestReturnArrayGuint8_2(t *testing.T) {
    Convey("ReturnArrayGuint8_2",t,func(){
        cvt := NewTypeConverter(true, "", emptyType, &mygi.Array{
            CType: "guchar*",
            ElemType: mygi.ElemType{
                Name: "guint8",
                CType: "guchar",
            },
        })
        // gotype []mygibase.Guchar
        //fmt.Println("ReturnArrayGuint8_2", cvt)
        So(cvt.GoType, ShouldEqual, "[]mygibase.Guchar")
        So(cvt.CgoType, ShouldEqual, "*C.guchar")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "mygibase.Guchar")
        So(cvt.ElemCgoType, ShouldEqual, "C.guchar")
    })
}
//<return-value transfer-ownership="full" nullable="1">
          //<doc xml:space="preserve">
 //a NUL-terminated byte array with the line that was read in
 //(without the newlines).  Set @length to a #gsize to get the length
 //of the read line.  On an error, it will return %NULL and @error
 //will be set. If there's no content to read, it will still return
 //%NULL, but @error won't be set.</doc>
          //<array c:type="char*">
            //<type name="guint8"/>
          //</array>
        //</return-value>
func TestReturnArrayGuint8_3(t *testing.T) {
    Convey("ReturnArrayGuint8_3", t, func(){
        cvt := NewTypeConverter(true, "", emptyType, &mygi.Array{
            CType: "char*",
            ElemType: mygi.ElemType{
                Name: "guint8",
                CType: "",
            },
        })
        // gotype []uint8
        //fmt.Println("ReturnArrayGuint8_3", cvt)
        So(cvt.GoType, ShouldEqual, "[]uint8")
        So(cvt.CgoType, ShouldEqual, "*C.char")
        So(cvt.IsArrayType, ShouldBeTrue)
        So(cvt.ElemGoType, ShouldEqual, "uint8")
        So(cvt.ElemCgoType, ShouldEqual, "C.char")
    })
}
        //<return-value transfer-ownership="full">
          //<doc xml:space="preserve">the list of
//CA DNs. You should unref each element with g_byte_array_unref() and then
//the free the list with g_list_free().</doc>
          //<type name="GLib.List" c:type="GList*">
            //<array name="GLib.ByteArray">
              //<type name="gpointer" c:type="gpointer"/>
            //</array>
          //</type>
        //</return-value>
func TestReturnGListArray(t *testing.T) {
    Convey("ReturnGListArray", t, func(){
        cvt := NewTypeConverter(true, "", &mygi.Type{
            Name: "GLib.List",
            CType: "GList*",
            Array: mygi.Array{
                Name: "GLib.ByteArray",
                ElemType: mygi.ElemType{
                    Name: "gpointer",
                    CType: "gpointer",
                },
            },
        },emptyArray)
        // gotype *glib.List
        //fmt.Println("ReturnGListArray", cvt)
        So(cvt.GoType, ShouldEqual, "*glib.List")
        So(cvt.CgoType, ShouldEqual, "*C.GList")
    })
}
//<return-value transfer-ownership="none">
          //<doc xml:space="preserve">%TRUE if @action_name is valid</doc>
          //<type name="gboolean" c:type="gboolean"/>
        //</return-value>
func TestReturnBool(t *testing.T) {
    Convey("ReturnReturnBool", t, func(){
        cvt := NewTypeConverter(true, "", &mygi.Type{
            Name: "gboolean",
            CType: "gboolean",
        }, emptyArray)
        // gotype bool
        //fmt.Println("ReturnBool", cvt)
        So(cvt.GoType, ShouldEqual, "bool")
        So(cvt.CgoType, ShouldEqual, "C.gboolean")
    })
}

//<parameter name="enabled"
                       //direction="out"
                       //caller-allocates="0"
                       //transfer-ownership="full">
              //<doc xml:space="preserve">if the action is presently enabled</doc>
              //<type name="gboolean" c:type="gboolean*"/>
            //</parameter>
func TestParamOutBool(t *testing.T) {
    Convey("ParamOutBool", t, func(){
        cvt := NewTypeConverter(false, "out", &mygi.Type{
            Name: "gboolean",
            CType: "gboolean*",
        }, emptyArray)
        // gotype bool
        //fmt.Println("ReturnBool", cvt)
        So(cvt.GoType, ShouldEqual, "bool")
        So(cvt.CgoType, ShouldEqual, "C.gboolean")
    })
}

//<parameter name="action_name" transfer-ownership="none">
            //<doc xml:space="preserve">an potential action name</doc>
            //<type name="utf8" c:type="const gchar*"/>
          //</parameter>
func TestParamString(t *testing.T) {
    Convey("ParamString", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "utf8",
            CType: "const gchar*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "string")
        So(cvt.CgoType, ShouldEqual, "*C.gchar")
    })
}

//<return-value transfer-ownership="full">
          //<doc xml:space="preserve">a detailed format string</doc>
          //<type name="utf8" c:type="gchar*"/>
        //</return-value>
func TestReturnString(t *testing.T) {
    Convey("ReturnString", t, func(){
        cvt := NewTypeConverter(true, "", &mygi.Type{
            Name: "utf8",
            CType: "gchar*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "string")
        So(cvt.CgoType, ShouldEqual, "*C.gchar")
    })
}

//<parameter name="action_name"
                     //direction="out"
                     //caller-allocates="0"
                     //transfer-ownership="full">
            //<doc xml:space="preserve">the action name</doc>
            //<type name="utf8" c:type="gchar**"/>
          //</parameter>
func TestParamOutString(t *testing.T) {
    Convey("ParamOutString", t, func(){
        cvt := NewTypeConverter(false, "out", &mygi.Type{
            Name: "utf8",
            CType: "gchar**",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "string")
        So(cvt.CgoType, ShouldEqual, "*C.gchar")
    })
}

//<parameter name="target_value"
                     //direction="out"
                     //caller-allocates="0"
                     //transfer-ownership="full">
            //<doc xml:space="preserve">the target value, or %NULL for no target</doc>
            //<type name="GLib.Variant" c:type="GVariant**"/>
          //</parameter>
func TestParamOutStruct(t *testing.T) {
    Convey("ParamOutStruct", t, func(){
        cvt := NewTypeConverter(false, "out", &mygi.Type{
            Name: "GLib.Variant",
            CType: "GVariant**",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*glib.Variant")
        So(cvt.CgoType, ShouldEqual, "*C.GVariant")
    })
}

//<parameter name="parameter_type"
           //direction="out"
           //caller-allocates="0"
           //transfer-ownership="full"
           //optional="1"
           //allow-none="1">
  //<doc xml:space="preserve">the parameter type, or %NULL if none needed</doc>
  //<type name="GLib.VariantType" c:type="const GVariantType**"/>
//</parameter>
func TestParamOutStruct1(t *testing.T) {
    Convey("ParamOutStruct1", t, func(){
        cvt := NewTypeConverter(false, "out", &mygi.Type{
            Name: "GLib.VariantType",
            CType: "const GVariantType**",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*glib.VariantType")
        So(cvt.CgoType, ShouldEqual, "*C.GVariantType")
    })
}

//<parameter name="target_value"
                     //transfer-ownership="none"
                     //nullable="1"
                     //allow-none="1">
            //<doc xml:space="preserve">a #GVariant target value, or %NULL</doc>
            //<type name="GLib.Variant" c:type="GVariant*"/>
          //</parameter>
func TestParamStruct(t *testing.T) {
    Convey("ParamStruct", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "GLib.Variant",
            CType: "GVariant*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*glib.Variant")
        So(cvt.CgoType, ShouldEqual, "*C.GVariant")
    })
}

//<instance-parameter name="action" transfer-ownership="none">
            //<doc xml:space="preserve">a #GAction</doc>
            //<type name="Action" c:type="GAction*"/>
          //</instance-parameter>
func TestParamStruct1(t *testing.T) {
    Convey("ParamStruct1", t, func(){
        cvt := NewTypeConverter(false, "", &mygi.Type{
            Name: "Action",
            CType: "GAction*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*Action")
        So(cvt.CgoType, ShouldEqual, "*C.GAction")
    })
}
        //<return-value transfer-ownership="none" nullable="1">
          //<doc xml:space="preserve">the parameter type</doc>
          //<type name="GLib.VariantType" c:type="const GVariantType*"/>
        //</return-value>
func TestReturnStruct2(t *testing.T) {
    Convey("ReturnStruct2", t, func(){
        cvt := NewTypeConverter(true, "", &mygi.Type{
            Name: "GLib.VariantType",
            CType: "const GVariantType*",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "*glib.VariantType")
        So(cvt.CgoType, ShouldEqual, "*C.GVariantType")
    })
}

//<return-value transfer-ownership="none">
          //<type name="none" c:type="void"/>
        //</return-value>
        //<parameters>
func TestNoReturn(t *testing.T) {
    Convey("NoReturn", t, func(){
        cvt := NewTypeConverter(true, "", &mygi.Type{
            Name: "none",
            CType: "void",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "none")
        So(cvt.CgoType, ShouldEqual, "C.void")
    })
}

func TestParamOutError(t *testing.T) {
    Convey("ParamOutError", t, func(){
        cvt := NewTypeConverter(false, "out", &mygi.Type{
            Name: "GError",
            CType: "GError**",
        }, emptyArray)
        So(cvt.GoType, ShouldEqual, "error")
        So(cvt.CgoType, ShouldEqual, "*C.GError")
    })
}
