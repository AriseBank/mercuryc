DOMAIN=apollo
POFILES=$(wildcard po/*.po)
MOFILES=$(patsubst %.po,%.mo,$(POFILES))
LINGUAS=$(basename $(POFILES))
POTFILE=po/$(DOMAIN).pot

# dist is primarily for use when packaging; for development we still manage
# dependencies via `go get` explicitly.
# TODO: use git describe for versioning
VERSION=$(shell grep "var Version" shared/version/flex.go | cut -d'"' -f2)
ARCHIVE=apollo-$(VERSION).tar
TAGS=$(shell test -e /usr/include/sqlite3.h && echo "-tags libsqlite3")

.PHONY: default
default:
	go get -t -v -d ./...
	go install -v $(TAGS) $(DEBUG) ./...
	@echo "APOLLO built successfully"

.PHONY: client
client:
	go get -t -v -d ./...
	go install -v $(TAGS) $(DEBUG) ./mercury
	@echo "APOLLO client built successfully"

.PHONY: update
update:
	go get -t -v -d -u ./...
	@echo "Dependencies updated"

.PHONY: debug
debug:
	go get -t -v -d ./...
	go install -v $(TAGS) -tags logdebug $(DEBUG) ./...
	@echo "APOLLO built successfully"

# This only needs to be done when migrate.proto is actually changed; since we
# commit the .pb.go in the tree and it's not expected to change very often,
# it's not a default build step.
.PHONY: protobuf
protobuf:
	protoc --go_out=. ./apollo/migrate.proto

.PHONY: check
check: default
	go get -v -x github.com/rogpeppe/godeps
	go get -v -x github.com/remyoudompheng/go-misc/deadcode
	go get -v -x github.com/golang/lint/golint
	go test -v $(TAGS) $(DEBUG) ./...
	cd test && ./main.sh

gccgo:
	go build -v $(TAGS) $(DEBUG) -compiler gccgo ./...
	@echo "APOLLO built successfully with gccgo"

.PHONY: dist
dist:
	# Cleanup
	rm -Rf $(ARCHIVE).gz

	# Create build dir
	$(eval TMP := $(shell mktemp -d))
	git archive --prefix=apollo-$(VERSION)/ HEAD | tar -x -C $(TMP)
	mkdir -p $(TMP)/dist/src/github.com/mercury
	ln -s ../../../../apollo-$(VERSION) $(TMP)/dist/src/github.com/mercury/apollo

	# Download dependencies
	cd $(TMP)/apollo-$(VERSION) && GOPATH=$(TMP)/dist go get -t -v -d ./...

	# Workaround for gorilla/mux on Go < 1.7
	cd $(TMP)/apollo-$(VERSION) && GOPATH=$(TMP)/dist go get -v -d github.com/gorilla/context

	# Assemble tarball
	rm $(TMP)/dist/src/github.com/mercury/apollo
	ln -s ../../../../ $(TMP)/dist/src/github.com/mercury/apollo
	mv $(TMP)/dist $(TMP)/apollo-$(VERSION)/
	tar --exclude-vcs -C $(TMP) -zcf $(ARCHIVE).gz apollo-$(VERSION)/

	# Cleanup
	rm -Rf $(TMP)

.PHONY: i18n update-po update-pot build-mo static-analysis
i18n: update-pot update-po

po/%.mo: po/%.po
	msgfmt --statistics -o $@ $<

po/%.po: po/$(DOMAIN).pot
	msgmerge -U po/$*.po po/$(DOMAIN).pot

update-po:
	for lang in $(LINGUAS); do\
	    msgmerge -U $$lang.po po/$(DOMAIN).pot; \
	    rm -f $$lang.po~; \
	done

update-pot:
	go get -v -x github.com/snapcore/snapd/i18n/xgettext-go/
	xgettext-go -o po/$(DOMAIN).pot --add-comments-tag=TRANSLATORS: --sort-output --package-name=$(DOMAIN) --msgid-bugs-address=mercury-devel@lists.linuxcontainers.org --keyword=i18n.G --keyword-plural=i18n.NG *.go shared/*.go mercury/*.go apollo/*.go

build-mo: $(MOFILES)

static-analysis:
	(cd test;  /bin/sh -x -c ". suites/static_analysis.sh; test_static_analysis")

tags: *.go apollo/*.go shared/*.go mercury/*.go
	find . | grep \.go | grep -v git | grep -v .swp | grep -v vagrant | xargs gotags > tags
