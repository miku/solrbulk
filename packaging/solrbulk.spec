Summary:    Fast parallel bulk loading utility for SOLR.
Name:       solrbulk
Version:    0.3.15
Release:    0
License:    MIT
BuildArch:  x86_64
BuildRoot:  %{_tmppath}/%{name}-build
Group:      System/Base
Vendor:     Leipzig University Library <http://ub.uni-leipzig.de>
URL:        https://github.com/miku/solrbulk

%description

Fast parallel bulk loading utility for SOLR.

%prep
# the set up macro unpacks the source bundle and changes in to the represented by
# %{name} which in this case would be my_maintenance_scripts. So your source bundle
# needs to have a top level directory inside called my_maintenance _scripts
# %setup -n %{name}

%build
# this section is empty for this example as we're not actually building anything

%install
# create directories where the files will be located
mkdir -p $RPM_BUILD_ROOT/usr/local/sbin

# put the files in to the relevant directories.
# the argument on -m is the permissions expressed as octal. (See chmod man page for details.)
install -m 755 solrbulk $RPM_BUILD_ROOT/usr/local/sbin

mkdir -p $RPM_BUILD_ROOT/usr/local/share/man/man1
install -m 644 solrbulk.1 $RPM_BUILD_ROOT/usr/local/share/man/man1/solrbulk.1

%post
# the post section is where you can run commands after the rpm is installed.
# insserv /etc/init.d/my_maintenance

%clean
rm -rf $RPM_BUILD_ROOT
rm -rf %{_tmppath}/%{name}
rm -rf %{_topdir}/BUILD/%{name}

# list files owned by the package here
%files
%defattr(-,root,root)
/usr/local/sbin/solrbulk
/usr/local/share/man/man1/solrbulk.1


%changelog
* Sat Nov 26 2016 Martin Czygan
- 0.2.0 release, fix bug that let solrbulk loose documents

* Mon Oct 26 2015 Martin Czygan
- 0.1.5.4 release, add -collection flag

* Tue Jan 20 2015 Martin Czygan
- 0.1.4 release
- autolimit solrbulk-tune test to number of lines in sample file
- remove -limit flag

* Tue Jan 20 2015 Martin Czygan
- 0.1.3 release
- added reset retry to solrbulk-tune

* Tue Jan 20 2015 Martin Czygan
- 0.1.2 release
- added solrbulk-tune, an experimental parameter optimizing script

* Mon Jan 19 2015 Martin Czygan
- 0.1.1 release
