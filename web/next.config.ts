import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
    output: "standalone", // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/output
    crossOrigin: 'anonymous', // TODO: ok? FIX?
    experimental: {
        serverActions: { // TODO; see https://nextjs.org/docs/app/api-reference/config/next-config-js/serverActions
            // TODO: go over these! figure out what is actually needed!
            allowedOrigins: ['appli.ng', 'mush.appli.ng', 'web', 'http://web:3000', 'web:3000', 'api'], // TODO: GET THESE FROM ELSEWHERE!
            // bodySizeLimit: '2mb', // TODO: ???
        },
        cssChunking: true, // TODO: ok?
        //inlineCss: true, // TODO: ????
        //optimizePackageImports: ['package-name'], // TODO: do this!
        // prefetchInlining: false, // TODO: consider via https://nextjs.org/docs/app/api-reference/config/next-config-js/prefetchInlining
        // prefetchInlining: { // TODO: consider via https://nextjs.org/docs/app/api-reference/config/next-config-js/prefetchInlining
        //     maxSize: 2048,
        //     maxBundleSize: 10240,
        // },
        //proxyClientMaxBodySize: 1048576, // 1MB in bytes // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/proxyClientMaxBodySize
        // staleTimes: { // Client caching of page segments // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/staleTimes
        //     dynamic: 30,
        //     static: 180,
        // },
        // TODO: check what to do (if anything) for the next 3! https://nextjs.org/docs/app/api-reference/config/next-config-js/staticGeneration
        // staticGenerationRetryCount: 1, // The number of times to retry a failed page generation before failing the build.
        // staticGenerationMaxConcurrency: 8, // The maximum number of pages to be processed per worker.
        // staticGenerationMinPagesPerWorker: 25, // The minimum number of pages to be processed before starting a new worker.
        //taint: true, // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/taint
        // turbopackChunking: { // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/turbopackChunking
        //     minChunkSize: 50000,
        //     maxChunkCountPerGroup: 40,
        //     maxMergeChunkSize: 200000,
        //     minComponentChunkSize: 20000,
        //     generateComponentChunks: false,
        // },
        turbopackFileSystemCacheForDev: true, // TODO: move to dev?
        //turbopackFileSystemCacheForBuild: true, // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/turbopackFileSystemCache
        //turbopackLocalPostcssConfig: true, // TODO: ???
        // turbopackMemoryEviction: 'auto', // TODO: ???
        // Use the Rust port instead of the Babel transform
        //turbopackRustReactCompiler: true, // TODO: must have reactCompiler: true,
        // urlImports: [ // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/urlImports
        // 'https://example.com/assets/', 'https://cdn.skypack.dev'
        // ],
        // TODO: useLightningcss: false, // default, ignored on Turbopack
        // useOffline: true, // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/useOffline
        // useTypeScriptCli: false,
        // webVitalsAttribution: ['CLS', 'LCP'], // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/webVitalsAttribution
    },
    trailingSlash: false, // Redirect trailing slashes to no trailing slash
    //typedRoutes: true, // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/typedRoutes
    //transpilePackages: ['package-name', '@scope/pkg'], // TODO: ???
    //reactCompiler: true, // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/reactCompiler
    // reactCompiler: {
    //     compilationMode: 'annotation',
    // },
    logging: {
        fetches: {
            fullUrl: true,
        },
        serverFunctions: true,
        // incomingRequests: { // TODO: this
        //     // ignore: [/\api\/v1\/health/],
        // },
        //incomingRequests: false, // TODO: or completely disable
        //browserToTerminal: true, // TODO: Forwards all browser logs to terminal 'warn', 'error', true, false
    },
    //serverExternalPackages: ['@acme/ui'], // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/serverExternalPackages
    cacheComponents: true, // TODO: ???
    //partialPrefetching: true, // TODO: https://nextjs.org/docs/app/api-reference/config/next-config-js/partialPrefetching
    // pageExtensions: ['js', 'jsx', 'ts', 'tsx', 'md', 'mdx'], // TODO: https://nextjs.org/docs/app/api-reference/config/next-config-js/pageExtensions
    // outputHashSalt: 'my-deployment-salt', // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/outputHashSalt
    // generateBuildId: async () => { // TODO: DO THIS!
    //     // This could be anything, using the latest git hash
    //     return process.env.GIT_HASH
    // },
    //deploymentId: 'my-deployment-id', // TODO: CHANGE?
    //generateEtags: false, // TODO: https://nextjs.org/docs/app/api-reference/config/next-config-js/generateEtags
    //expireTime: 3600, // one hour in seconds // TODO: how long for CDNs to cache pages....
    // distDir: 'build', // TODO: uses /build instead of /.next
    // compress: false, // TODO: DELETE
    // dev: true, // TODO: DELETE
    // devIndicators: { // TODO: DELETE
    //     position: 'bottom-right', // 'bottom-left' | 'bottom-right' | 'top-left' | 'top-right'
    // },
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
    //htmlLimitedBots: /MySpecialBot|MyAnotherSpecialBot|SimpleCrawler/, // TODO: https://nextjs.org/docs/app/api-reference/config/next-config-js/htmlLimitedBots
    // httpAgentOptions: { // TODO: ??? https://nextjs.org/docs/app/api-reference/config/next-config-js/httpAgentOptions
    //     keepAlive: false,
    // },
    // images: { // TODO: ???? https://nextjs.org/docs/app/api-reference/config/next-config-js/images
    //     loader: 'custom',
    //     loaderFile: './my/image/loader.js',
    // },
    // cacheHandler: require.resolve('./cache-handler.js'), // TODO: ??? https://nextjs.org/docs/app/api-reference/config/next-config-js/incrementalCacheHandlerPath
    // cacheMaxMemorySize: 0, // disable default in-memory caching
    // turbopack: { // TODO: may not need, but see https://nextjs.org/docs/app/api-reference/config/next-config-js/turbopack
    //      ignoreIssue: [
    //       {
    //         path: '**/vendor/**',
    //       },
    //     ],
    //     // TODO: ???
    // },

    poweredByHeader: false,
    // reactMaxHeadersLength: 1000, // Default 6000 // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/reactMaxHeadersLength
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
            // { // TODO: the login path!
            //     // matching all API routes
            //     source: ":path*",
            //     headers: [
            //         { key: "Access-Control-Allow-Credentials", value: "true" },
            //         { key: "Access-Control-Allow-Origin", value: "*" }, // replace this your actual origin // TODO: fix!
            //         { key: "Access-Control-Allow-Methods", value: "GET,DELETE,PATCH,POST,PUT" },
            //         { key: "Access-Control-Allow-Headers", value: "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version" },
            //     ]
            // }
        ]
    },
    redirects() {
        return [ // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/redirects
            // {
            //     source: '/about',
            //     destination: '/',
            //     permanent: true,
            // },
            // {
            //     source: '/old-blog/:path*',
            //     destination: '/blog/:path*',
            //     permanent: false
            // }
        ]
    },
    rewrites() {
        return [ // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/rewrites
            // {
            //     source: '/about',
            //     destination: '/',
            // },
        ]
    },
};
// // Dev exports
// module.exports = {
// typescript: { // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/typescript
//     ignoreBuildErrors: false,
//     tsconfigPath: 'tsconfig.json',
// },
//     productionBrowserSourceMaps: true,
//     output: "standalone",
// };

// Prod Exports
module.exports = {
    // typescript: { // TODO: see https://nextjs.org/docs/app/api-reference/config/next-config-js/typescript
    //     ignoreBuildErrors: false,
    //     tsconfigPath: 'tsconfig.json',
    // },
    productionBrowserSourceMaps: false, // TODO: only enable for debugging! Allows console debugging...
    output: "standalone",
};

export default nextConfig;


