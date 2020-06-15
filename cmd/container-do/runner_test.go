package main

import (
    "reflect"
    "testing"
)

func Test_parseOsReleaseFile(t *testing.T) {
    type args struct {
        data []byte
    }
    tests := []struct {
        name    string
        args    args
        want    map[string]string
        wantErr bool
    }{
        {
            name: "simple file",
            args: args{
                []byte("FOO=\"BAR ISTA\"\nbar=foo"),
            },
            want: map[string]string{
                "FOO": "BAR ISTA",
                "bar": "foo",
            },
            wantErr: false,
        },
        {
            name: "ignore empty line",
            args: args{
                []byte("FOO=\"BAR\"\n\nbar=foo"),
            },
            want: map[string]string{
                "FOO": "BAR",
                "bar": "foo",
            },
            wantErr: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseOsReleaseFile(tt.args.data)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseOsReleaseFile() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("parseOsReleaseFile() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func Test_extractOsFlavorFromReleaseFile(t *testing.T) {
    type args struct {
        out []byte
    }
    tests := []struct {
        name    string
        args    args
        want    string
        wantErr bool
    }{
        {
            name: "debian",
            args: args{
                []byte(`PRETTY_NAME="Debian GNU/Linux 10 (buster)"
NAME="Debian GNU/Linux"
VERSION_ID="10"
VERSION="10 (buster)"
VERSION_CODENAME=buster
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"`),
            },
            want:    "debian",
            wantErr: false,
        },
        {
            name: "debian",
            args: args{
                []byte(`PRETTY_NAME="Debian GNU/Linux 10 (buster)"
NAME="Debian GNU/Linux"
VERSION_ID="10"
VERSION="10 (buster)"
VERSION_CODENAME=buster
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"`),
            },
            want:    "debian",
            wantErr: false,
        },
        {
            name: "ubuntu",
            args: args{
                []byte(`NAME="Ubuntu"
VERSION="20.04 LTS (Focal Fossa)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 20.04 LTS"
VERSION_ID="20.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=focal
UBUNTU_CODENAME=focal`),
            },
            want:    "debian",
            wantErr: false,
        },
        {
            name: "rhel8",
            args: args{
                []byte(`NAME="Red Hat Enterprise Linux"
VERSION="8.2 (Ootpa)"
ID="rhel"
ID_LIKE="fedora"
VERSION_ID="8.2"
PLATFORM_ID="platform:el8"
PRETTY_NAME="Red Hat Enterprise Linux 8.2 (Ootpa)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:redhat:enterprise_linux:8.2:GA"
HOME_URL="https://www.redhat.com/"
BUG_REPORT_URL="https://bugzilla.redhat.com/"

REDHAT_BUGZILLA_PRODUCT="Red Hat Enterprise Linux 8"
REDHAT_BUGZILLA_PRODUCT_VERSION=8.2
REDHAT_SUPPORT_PRODUCT="Red Hat Enterprise Linux"
REDHAT_SUPPORT_PRODUCT_VERSION="8.2"`),
            },
            want:    "fedora",
            wantErr: false,
        },
        {
            name: "centos",
            args: args{
                []byte(`NAME="CentOS Linux"
VERSION="8 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="8"
PLATFORM_ID="platform:el8"
PRETTY_NAME="CentOS Linux 8 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:8"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-8"
CENTOS_MANTISBT_PROJECT_VERSION="8"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="8"`),
            },
            want:    "fedora",
            wantErr: false,
        },
        {
            name: "alpine",
            args: args{
                []byte(`NAME="Alpine Linux"
ID=alpine
VERSION_ID=3.12.0
PRETTY_NAME="Alpine Linux v3.12"
HOME_URL="https://alpinelinux.org/"
BUG_REPORT_URL="https://bugs.alpinelinux.org/"
`),
            },
            want:    "alpine",
            wantErr: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := extractOsFlavorFromReleaseFile(tt.args.out)
            if (err != nil) != tt.wantErr {
                t.Errorf("extractOsFlavorFromReleaseFile() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("extractOsFlavorFromReleaseFile() got = %v, want %v", got, tt.want)
            }
        })
    }
}
