Name:       versaDumps
Version:    0.0.0
Release:    1%{?dist}
Summary:    VersaDumps Visualizer

License:    MIT
URL:        https://github.com/kriollo/versaDumps
Source0:    %{name}-%{version}-linux-amd64.tar.gz

BuildArch:  x86_64
Requires:   gtk3, libwebp, libX11

%description
VersaDumps Visualizer - monitor de dumps y logs.

%prep

%build

%install
mkdir -p %{buildroot}/usr/bin
install -m 0755 versaDumps %{buildroot}/usr/bin/versaDumps
mkdir -p %{buildroot}/usr/share/icons/hicolor/256x256/apps
install -m 0644 app/build/linux/icon.png %{buildroot}/usr/share/icons/hicolor/256x256/apps/versaDumps.png
mkdir -p %{buildroot}/usr/share/applications
install -m 0644 packaging/fedora/versaDumps.desktop %{buildroot}/usr/share/applications/versaDumps.desktop

%files
/usr/bin/versaDumps
/usr/share/icons/hicolor/256x256/apps/versaDumps.png
/usr/share/applications/versaDumps.desktop

%changelog
* Tue Feb 14 2026 Your Name - 0.0.0-1
- Initial Fedora packaging
