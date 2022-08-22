import { Engine, GraphQLClient, gql } from "@dagger.io/dagger";

new Engine({
  LocalDirs: {
    app: ".",
  },
  Port: 9999,
}).run(async (client: GraphQLClient) => {
  const output = await client.request(
    gql`
      {
        core {
          clientdir(id: "app") {
            file(path: "cloak.yaml")
          }
        }
      }
    `
  );
  console.log(output);
});
