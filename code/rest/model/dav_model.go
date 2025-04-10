package model

import (
	"box/code/tool/dav"
	"box/code/tool/dav/xml"
	"fmt"
	"net/http"
	"path"
	"strconv"
)

// webdav url prefix.
var WEBDAV_PREFIX = "/api/dav"

// live prop.
type LiveProp struct {
	FindFn func(space *Space, matter *Matter) string
	Dir    bool
}

// all live prop map.
var LivePropMap = map[xml.Name]LiveProp{
	{Space: "DAV:", Local: "resourcetype"}: {
		FindFn: func(space *Space, matter *Matter) string {
			if matter.Dir {
				return `<D:collection xmlns:D="DAV:"/>`
			} else {
				return ""
			}
		},
		Dir: true,
	},
	{Space: "DAV:", Local: "displayname"}: {
		FindFn: func(space *Space, matter *Matter) string {
			if path.Clean("/"+matter.Name) == "/" {
				return ""
			} else {
				return dav.EscapeXML(matter.Name)
			}
		},
		Dir: true,
	},
	{Space: "DAV:", Local: "getcontentlength"}: {
		FindFn: func(space *Space, matter *Matter) string {
			return strconv.FormatInt(matter.Size, 10)
		},
		Dir: false,
	},
	{Space: "DAV:", Local: "getlastmodified"}: {
		FindFn: func(space *Space, matter *Matter) string {
			return matter.UpdateTime.UTC().Format(http.TimeFormat)
		},
		// http://webdav.org/specs/rfc4918.html#PROPERTY_getlastmodified
		// suggests that getlastmodified should only apply to GETable
		// resources, and this package does not support GET on directories.
		//
		// Nonetheless, some WebDAV clients expect child directories to be
		// sortable by getlastmodified date, so this value is true, not false.
		// See golang.org/issue/15334.
		Dir: true,
	},
	{Space: "DAV:", Local: "creationdate"}: {
		FindFn: nil,
		Dir:    false,
	},
	{Space: "DAV:", Local: "getcontentlanguage"}: {
		FindFn: nil,
		Dir:    false,
	},
	{Space: "DAV:", Local: "getcontenttype"}: {
		FindFn: func(space *Space, matter *Matter) string {
			if matter.Dir {
				return ""
			} else {
				return dav.EscapeXML(matter.Name)
			}
		},
		Dir: false,
	},
	{Space: "DAV:", Local: "getetag"}: {
		FindFn: func(space *Space, matter *Matter) string {
			return fmt.Sprintf(`"%x%x"`, matter.UpdateTime.UnixNano(), matter.Size)
		},
		// findETag implements ETag as the concatenated hex values of a file's
		// modification time and size. This is not a reliable synchronization
		// mechanism for directories, so we do not advertise getetag for DAV
		// collections.
		Dir: false,
	},
	// TODO: The lockdiscovery property requires LockSystem to list the
	// active locks on a resource.
	{Space: "DAV:", Local: "lockdiscovery"}: {},
	{Space: "DAV:", Local: "supportedlock"}: {
		FindFn: func(space *Space, matter *Matter) string {
			return `` +
				`<D:lockentry xmlns:D="DAV:">` +
				`<D:lockscope><D:exclusive/></D:lockscope>` +
				`<D:locktype><D:write/></D:locktype>` +
				`</D:lockentry>`
		},
		Dir: true,
	},
	{Space: "DAV:", Local: "quota-available-bytes"}: {
		FindFn: func(space *Space, matter *Matter) string {
			var size int64 = 0
			if space.TotalSizeLimit >= 0 {
				if space.TotalSizeLimit-space.TotalSize > 0 {
					size = space.TotalSizeLimit - space.TotalSize
				} else {
					size = 0
				}
			} else {
				// no limit, default 100G.
				size = 100 * 1024 * 1024 * 1024
			}
			return fmt.Sprintf(`%d`, size)
		},
		Dir: true,
	},
	{Space: "DAV:", Local: "quota-used-bytes"}: {
		FindFn: func(space *Space, matter *Matter) string {
			return fmt.Sprintf(`%d`, space.TotalSize)
		},
		Dir: true,
	},
}
