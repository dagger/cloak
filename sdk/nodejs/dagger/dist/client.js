import axios from "axios";
import { buildAxiosFetch } from "@lifeomic/axios-fetch";
import { GraphQLClient } from "graphql-request";
import { Response } from "node-fetch";
// @ts-expect-error node-fetch doesn't exactly match the Response object, but close enough.
global.Response = Response;
export const client = new GraphQLClient("http://fake.invalid/graphql", {
    fetch: buildAxiosFetch(axios.create({
        socketPath: "/dagger.sock",
        timeout: 3600e3,
    })),
});
export class Client {
    constructor() {
        this.client = axios.create({
            socketPath: "/dagger.sock",
            timeout: 3600e3,
        });
    }
    async do(payload) {
        const response = await this.client.post(`http://fake.invalid/graphql`, payload, { headers: { "Content-Type": "application/graphql" } });
        return response;
    }
}
export class FSID {
    constructor(serial) {
        this.serial = serial;
    }
    toString() {
        return this.serial;
    }
    toJSON() {
        return this.serial;
    }
}
export class SecretID {
    constructor(serial) {
        this.serial = serial;
    }
    toString() {
        return this.serial;
    }
    toJSON() {
        return this.serial;
    }
}
//# sourceMappingURL=client.js.map