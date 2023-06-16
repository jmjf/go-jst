ARG VARIANT=1.20-bullseye
FROM golang:${VARIANT}

ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=${USER_UID}
RUN groupadd --gid ${USER_GID} ${USERNAME} \
   && useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} \
   && chsh -s /bin/bash ${USERNAME} \
   && apt-get update \
   && apt-get install -y sudo git \
   && echo ${USERNAME} ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/${USERNAME} \
   && chmod 0440 /etc/sudoers.d/${USERNAME}

ARG NODE_VERSION="lts/*"
RUN if [ "${NODE_VERSION}" != "none" ]; then su dev -c "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash"; fi

