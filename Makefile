default: generate

# Runs the ego templating generation tool whenever an HTML template changes.
generate: api/html/*.ego
	@go run github.com/benbjohnson/ego/cmd/ego ./api/html

# Removes all ego Go files from the http/html directory.
clean:
	@rm api/html/*.ego.go

# Removes the third party theme from the file system.
remove-theme:
	@rm api/assets/css/theme.css

.PHONY: default generate clean remove-theme