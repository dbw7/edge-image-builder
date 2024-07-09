# ----- EIB Builder Image -----
FROM registry.suse.com/bci/golang:1.22

# Dependency uses by line
# 1. Podman Go library
RUN zypper install -y \
    gpgme-devel device-mapper-devel libbtrfs-devel

WORKDIR /src

COPY go.mod go.sum ./
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY .git .git

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    go build ./cmd/eib

# ----- Deliverable Image -----
FROM opensuse/leap:15.5

# Dependency uses by line
# 1. ISO image building
# 2. RAW image modification on x86_64
# 3. Podman EIB library
# 4. RPM resolution logic
# 5. Embedded artefact registry
# 6. Network configuration
RUN zypper addrepo https://download.opensuse.org/repositories/isv:SUSE:Edge:EdgeImageBuilder/SLE-15-SP5/isv:SUSE:Edge:EdgeImageBuilder.repo && \
    zypper --gpg-auto-import-keys refresh && \
    zypper install -y \
    xorriso squashfs  \
    libguestfs kernel-default e2fsprogs parted gptfdisk btrfsprogs guestfs-tools lvm2 \
    podman \
    createrepo_c \
    helm hauler \
    nm-configurator && \
    zypper clean -a

# Make adjustments for running guestfish and image modifications on aarch64
# guestfish looks for very specific locations on the filesystem for UEFI firmware
# and also expects the boot kernel to be a portable executable (PE), not ELF.
RUN zypper install -y qemu-uefi-aarch64 && \
	mkdir -p /usr/share/edk2/aarch64 && \
	cp /usr/share/qemu/aavmf-aarch64-code.bin /usr/share/edk2/aarch64/QEMU_EFI-pflash.raw && \
	cp /usr/share/qemu/aavmf-aarch64-vars.bin /usr/share/edk2/aarch64/vars-template-pflash.raw && \
	mv /boot/vmlinux* /boot/backup-vmlinux && \
	zypper clean -a

COPY --from=0 /src/eib /bin/eib

ENTRYPOINT ["/bin/eib"]
