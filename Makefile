PREFIX ?= /usr
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR ?= /etc
BINARY = batctl
MODULE = github.com/spaceclam/batctl

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -s -w -X main.version=$(VERSION)

.PHONY: all build install uninstall clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/batctl/

install: build
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 configs/batctl.service $(DESTDIR)$(SYSCONFDIR)/systemd/system/batctl.service
	install -Dm644 configs/99-batctl-resume.rules $(DESTDIR)$(SYSCONFDIR)/udev/rules.d/99-batctl-resume.rules

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	rm -f $(DESTDIR)$(SYSCONFDIR)/systemd/system/batctl.service
	rm -f $(DESTDIR)$(SYSCONFDIR)/udev/rules.d/99-batctl-resume.rules
	rm -f $(DESTDIR)$(SYSCONFDIR)/batctl.conf

clean:
	rm -f $(BINARY)
