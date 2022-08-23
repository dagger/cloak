import { Engine, gql } from "@dagger.io/dagger";
import { dirname } from "path";
import { fileURLToPath } from "url";
const __dirname = dirname(fileURLToPath(import.meta.url));

new Engine({
  ConfigDir: __dirname,
  LocalDirs: {
    app: ".",
  },
}).run(async (client) => {
  const output = await client.request(
    gql`
      {
        core {
          clientdir(id: "app") {
            file(path: "package.json")
          }
        }
      }
    `
  );
  console.log(output);
});
