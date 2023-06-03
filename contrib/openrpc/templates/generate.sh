#!/bin/sh
cat types.gotmpl | sed 's/`/`\+"`"\+`/g' | sed '1s/^/package templates \nconst TEMPLATE=`/' | sed '$a`' > template.go
