# Java Project with Maven

Commands to try:

```bash
container-do mvn test
container-do mvn exec:java -Dexec.mainClass="my.app.App"
```

_Note:_ By its nature, Maven downloads quite a few dependency on its first run,
here triggered in `run.setup`.
You may want to follow the 
  [official documentation](https://hub.docker.com/_/maven)
and persist the local repository in a volume:

```bash
docker volume create --name container-do-java-example-maven-repo
```

And in `ContainerDo.toml`:

```toml
[container]
mounts = [
    ".:/app",
    "container-do-java-example-maven-repo:/root/.m2"
]
```

Note that we have to make the working-directory mount explicit now!
