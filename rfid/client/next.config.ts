import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
    output: "standalone",
    crossOrigin: 'anonymous', // TODO: ok?
    experimental: {
        serverActions: {
            allowedOrigins: ['appli.ng', 'mush.appli.ng', 'web', 'http://web:3000', 'web:3000', 'api'],
        },
    },

    // compress: false, // TODO: DELETE
    // dev: true, // TODO: DELETE
    // webpack: (config, { isServer }) => {
    //     config.optimization.minimizer = [];// TODO: change back to true!
    //     if (!isServer) {
    //         config.optimization.minimize = false; // TODO: change back to true!
    //         // config.resolve.fallback = {
    //         //     ...config.resolve.fallback,
    //         //     perf_hooks: false,
    //         //     worker_threads: false,
    //         // };
    //     }
    //     config.optimization.minimize = false; // TODO: change back to true!
    //     return config;
    // },
    async headers() {
        return [
            {
                // matching all API routes
                source: ":path*",
                headers: [
                    { key: "Access-Control-Allow-Credentials", value: "true" },
                    { key: "Access-Control-Allow-Origin", value: "*" }, // replace this your actual origin // TODO: fix!
                    { key: "Access-Control-Allow-Methods", value: "GET,DELETE,PATCH,POST,PUT" },
                    { key: "Access-Control-Allow-Headers", value: "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version" },
                ]
            }
        ]
    }
};

module.exports = {
    productionBrowserSourceMaps: true, // TODO: deleteme
    output: "standalone",
};

export default nextConfig;


