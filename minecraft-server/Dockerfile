FROM eclipse-temurin:25-jdk-jammy

WORKDIR /minecraft

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -u 1000 mc

# 将已安装好的 Fabric 服务端文件放入构建上下文（含 fabric-server-launch.jar、libraries、mods 等）
COPY --chown=mc:mc . .
RUN chown -R mc:mc /minecraft

ENV EULA=FALSE
ENV MEMORY=2G
ENV JAVA_OPTS="-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200"

EXPOSE 25565

VOLUME ["/minecraft/world", "/minecraft/mods", "/minecraft/config"]

USER mc

CMD ["bash", "-c", "if [ \"$EULA\" != \"TRUE\" ]; then echo '请设置环境变量 EULA=TRUE 以接受 Minecraft EULA'; exit 1; fi; echo eula=true > eula.txt; exec java -Xms${MEMORY} -Xmx${MEMORY} ${JAVA_OPTS} -jar fabric-server-launch.jar nogui"]
