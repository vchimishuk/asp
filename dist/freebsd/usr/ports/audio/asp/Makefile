PORTNAME=	asp
PORTVERSION=	0.0.1
CATEGORIES=	audio

MAINTAINER=	vchimishuk@yandex.ru
COMMENT=	ncurses client for Chub audio player
WWW=		https://github.com/vchimishuk/asp

LICENSE=	GPLv3+

FETCH_DEPENDS=	git:devel/git
USES=		go:1.21,modules

USE_GITHUB=	yes
GH_ACCOUNT=	vchimishuk
GH_PROJECT=	asp
GH_TAGNAME=	451d4a1
# $ make gomod-vendor
GH_TUPLE=	\
		gbin:goncurses:be6a464272ae:gbin_goncurses/vendor/github.com/gbin/goncurses \
		vchimishuk:chubby:94b95014d5a6:vchimishuk_chubby/vendor/github.com/vchimishuk/chubby \
		vchimishuk:config:ed7bd1b64558:vchimishuk_config/vendor/github.com/vchimishuk/config \
		vchimishuk:opt:e09d5e511662:vchimishuk_opt/vendor/github.com/vchimishuk/opt

PLIST_FILES=	bin/asp

.include <bsd.port.mk>
