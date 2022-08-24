import { Engine, gql } from "@dagger.io/dagger";
import { dirname } from "path";
import { fileURLToPath } from "url";
const __dirname = dirname(fileURLToPath(import.meta.url));

new Engine({
  ConfigDir: __dirname,
}).run(async (client) => {
  const workdir = await client
    .request(
      gql`
        {
          host {
            workdir {
              read {
                id
              }
            }
          }
        }
      `
    )
    .then((res) => res.host.workdir.read.id);

  const build = await client
    .request(
      gql`
			{
				yarn {
					script(source: "${workdir}", name: "build") {
						id
					}
				}
			}
			`
    )
    .then((result) => result.yarn.script.id);

  await client.request(
    gql`
      {
        host {
          workdir {
            write(contents: "${build}")
          }
        }
      }
    `
  );
});
