#!/bin/sh -ex
cmd=./try-error.sh
$cmd glib-2.0
$cmd gobject-2.0
$cmd gio-2.0
$cmd gudev-1.0
$cmd flatpak-1.0
$cmd gdkpixbuf-2.0
$cmd pango-1.0
$cmd cairo-1.0
$cmd atk-1.0
$cmd gdk-3.0
#$cmd gtk-3.0

