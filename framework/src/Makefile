include $(GOROOT)/src/Make.inc

DIRS=\
pkg/util/dlog\
pkg/util/osutil\
pkg/util/pipes\
pkg/util/ini\
pkg/util/json\
pkg/util/iomod\
pkg/util/time\
pkg/util/session\
pkg/util/user\
pkg/util/perms\
pkg/util/paths\
pkg/util/fnotify\
pkg/util/template\
pkg/util/openid\
pkg/util/picasa\
pkg/util/pages\
pkg/util/moduled\
wfdr\
wfdr-deamon\
wfdr-cache-monitor\
wfdr-module-manager\
css-preprocessor\

all: install

clean.dirs: $(addsuffix .clean, $(DIRS))

install.dirs: $(addsuffix .install, $(DIRS))

%.clean:
	+cd $* && gomake clean
	
%.install:
	+cd $* && gomake install
	
clean: clean.dirs

install: install.dirs
