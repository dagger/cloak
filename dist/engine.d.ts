import { GraphQLClient } from "graphql-request";
export interface EngineOptions {
    ConfigDir: string;
    LocalDirs?: Record<string, string>;
    Port?: number;
}
export declare class Engine {
    private config;
    constructor(config: EngineOptions);
    run(cb: (client: GraphQLClient) => Promise<void>): Promise<void>;
}
//# sourceMappingURL=engine.d.ts.map