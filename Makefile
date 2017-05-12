.PHONY: up

link:
	rm -rf ./vendor/github.com/pinshare/{config,spec,syncker}
	cd vendor/github.com/pinshare && \
		ln -s ../../../../../../../../config/src/github.com/pinshare/config ./ && \
		ln -s ../../../../../../../../spec ./ && \
		ln -s ../../../../../../../../syncker/src/github.com/pinshare/syncker ./

up:
	glide up

uplink: up link
