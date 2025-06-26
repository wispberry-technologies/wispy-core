#!/bin/bash

# Set the Go packages location explicitly
GO_PACKAGES="/opt/homebrew/Cellar/go/1.24.2/libexec/src"

# Define the source and destination directories
SRC_TEXT_TEMPLATE="$GO_PACKAGES/text/template"
SRC_HTML_TEMPLATE="$GO_PACKAGES/html/template"
DEST_TEMPLATE="./pkg/"


# Check if the source directories exist
if [ ! -d "$SRC_TEXT_TEMPLATE" ]; then
  echo "Error: Source directory '$SRC_TEXT_TEMPLATE' not found."
  exit 1
fi

if [ ! -d "$SRC_HTML_TEMPLATE" ]; then
  echo "Error: Source directory '$SRC_HTML_TEMPLATE' not found."
  exit 1
fi

# Create the destination directory if it doesn't exist
mkdir -p "$DEST_TEMPLATE"

# Copy the text/template package
cp -r "$SRC_TEXT_TEMPLATE" "$DEST_TEMPLATE/textTemplate"

# Copy the html/template package
cp -r "$SRC_HTML_TEMPLATE" "$DEST_TEMPLATE/htmlTemplate"

echo "Successfully copied text/template to pkg/textTemplate & html/template to pkg/htmlTemplate"

# Function to replace imports in all Go files
replace_imports() {
  local dir="$1"
  
  echo "Replacing imports in $dir"
  
  # Find all .go files recursively in the directory
  find "$dir" -name "*.go" -type f | while read -r file; do
    echo "Processing $file"
    
    # Replace the imports
    sed -i '' \
      -e 's#"text/template/parse"#"wispy-core/pkg/textTemplate/parse"#g' \
      -e 's#"text/template"#"wispy-core/pkg/textTemplate"#g' \
      -e 's#"html/template"#"wispy-core/pkg/htmlTemplate"#g' \
      # -e 's#"text/template/#"github.com/gohugoio/hugo/tpl/internal/go_templates/texttemplate/#g' \
      # -e 's#"internal/fmtsort"#"github.com/gohugoio/hugo/tpl/internal/go_templates/fmtsort"#g' \
      # -e 's#"internal/testenv"#"github.com/gohugoio/hugo/tpl/internal/go_templates/testenv"#g' \
      "$file"
  done
}

# Replace imports in both packages
replace_imports "$DEST_TEMPLATE/htmlTemplate"

echo "Successfully replaced imports in template packages"
echo "Import replacements:"
echo "  - \"text/template\" → \"wispy-core/pkg/textTemplate\""
echo "  - \"html/template\" → \"wispy-core/pkg/htmlTemplate\""
# echo "  - \"text/template/\" → \"github.com/gohugoio/hugo/tpl/internal/go_templates/texttemplate/\""
# echo "  - \"internal/fmtsort\" → \"github.com/gohugoio/hugo/tpl/internal/go_templates/fmtsort\""
# echo "  - \"internal/testenv\" → \"github.com/gohugoio/hugo/tpl/internal/go_templates/testenv\""