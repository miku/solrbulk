NAME = solrbulk
VERSION = 0.3.9
RELEASE = 0

$(NAME)-$(VERSION)-$(RELEASE)-x86_64.pkg.tar.xz: PKGBUILD
	makepkg

clean:
	rm -f $(NAME)-$(VERSION)-$(RELEASE)-x86_64.pkg.tar.xz
	rm -f v$(VERSION).tar.gz
	rm -rf pkg/
	rm -rf src/
