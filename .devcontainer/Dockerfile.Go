FROM fedora:42 AS dev-container

# nvim stuff
RUN dnf install -y \
        gcc \
        cascadia-mono-nf-fonts \
        git \
        make \ 
        npm \ 
        nvim \ 
        ripgrep \ 
        shadow-utils \ 
        unzip \ 
        jetbrains-mono-fonts
# Go stuff
RUN dnf install -y \
        golang
# C++ stuff
RUN dnf install -y \
        wayland-devel \
        libxkbcommon-devel \
        libX11-devel \
        libXext-devel \
        libXrandr-devel \
        libXinerama-devel \
        libXcursor-devel \
        libXi-devel \
        libXxf86vm-devel \
        libXScrnSaver-devel \
        mesa-libGL-devel \
        mesa-libGLU-devel \
        cmake \
        clang \
        clang-tools-extra && \
dnf copr enable -y che/nerd-fonts && dnf install -y nerd-fonts && \
dnf clean all

RUN \
go install github.com/air-verse/air@v1.62.0 && \
go install github.com/go-delve/delve/cmd/dlv@v1.7.3 && \
go install golang.org/x/tools/gopls@latest

RUN git clone --branch mine https://github.com/douglascdev/kickstart.nvim /root/.config/nvim

WORKDIR /usr/src

EXPOSE 8080
