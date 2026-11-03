import type {MetadataRoute} from 'next'
import {BaseExternalUrl} from "@/app/components/Constants";

// TODO: use these maybe once fixed!
export async function generateSitemaps() {
    // Fetch the total number of products and calculate the number of sitemaps needed
    // TODO: FIX!
    return [{id: 0}, {id: 1}, {id: 2}, {id: 3}]
}

// TODO: use these maybe once fixed!
export default async function sitemap(props: {
    id: Promise<string>
}): Promise<MetadataRoute.Sitemap> {
    // TODO: FIX!
    const id = await props.id
    // Google's limit is 50,000 URLs per sitemap
    //const start = id * 50000
    //const end = start + 50000
    // const products = await getProducts(
    //     `SELECT id, date FROM products WHERE id BETWEEN ${start} AND ${end}`
    // )
    const listTyps = ["agarBatches", "agarRecipes", "bags", "fruits", "fruitingChambers", "grainBatches", "jars", "jarRecipes", "lcs", "lcRecipes", "lcSyringes", "mss", "pcRuns", "plates", "plugs", "projects", "sales", "slants", "species", "sporePrints", "sporeSwabs", "stasisTubes", "subspecies", "substrateBatches", "substrateRecipes", "transfers", "users", "waterJars"]
    const importTyps = ["bag", "fruit", "fruitingChamber", "jar", "lc", "lcSyringe", "mss", "plate", "plugs", "slant", "sporePrint", "sporeSwab", "stasisTube", "waterJar"]
    const newTyps = ["agarRecipe", "jarRecipe", "lcRecipe", "pcRun", "plugs", "project", "species", "subspecies", "substrateRecipe", "waterJar"]
    const policiesPages = ["cookies","privacy","security"] // TODO: ensure list is complete
    const staticPoliciesPaths = ["cookies","privacy","security"].map(f=>{return `policies/${f}.txt`})
    const staticContentPaths = ["securityAcknowledgements.txt",...staticPoliciesPaths].map(p=>{
        return `staticContent/${p}`
    })
    const publicFiles = ["security.txt", "robots.txt"/* TODO: ADD!*/, ...staticContentPaths] // TODO: ensure list is complete


    return [
        { // Main page
            url: BaseExternalUrl || "mush.appli.ng",
            changeFrequency: "never",
            //lastModified: // TODO: ?????
            priority: 1.0
            // alternates?: { // TODO: ???
            //     languages?: Languages<string> | undefined
            // } | undefined,
            // images?: string[] | undefined,// TODO: ???
            // videos?: Videos[] | undefined,// TODO: ???
        },
        { // Login page
            url:  `${BaseExternalUrl}/login`,
            changeFrequency: "never",
            //lastModified: // TODO: ?????
            priority: 0.5 // TODO: ?????
            // alternates?: { // TODO: ???
            //     languages?: Languages<string> | undefined
            // } | undefined,
            // images?: string[] | undefined,// TODO: ???
            // videos?: Videos[] | undefined,// TODO: ???
        },
        // List pages
        ...listTyps.map(typ => {
            return {
                url: `${BaseExternalUrl}/list/${typ}`,
                //changeFrequency: "always" // TODO: ???
                //lastModified: // TODO: ?????
                priority: 0.6 // TODO: ?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
                // images?: string[] | undefined,// TODO: ???
                // videos?: Videos[] | undefined,// TODO: ???
            }
        }),
        // Import pages
        ...importTyps.map(typ => {
            return {
                url: `${BaseExternalUrl}/import/${typ}`,
                changeFrequency: "never",
                //lastModified: // TODO: ?????
                priority: 0.6 // TODO: ?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
            }
        }),
        // Creation pages
        ...newTyps.map(typ => {
            return {
                url: `${BaseExternalUrl}/new/${typ}`,
                changeFrequency: "never",
                //lastModified: // TODO: ?????
                priority: 0.5 // TODO: ?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
            }
        }),
        // TODO: view pages? (must generate at the beginning, probably dont do this)
        // policies pages (cookies, privacy)
        ...policiesPages.map(policyName => {
            return {
                url: `${BaseExternalUrl}/policies/${policyName}`,
                changeFrequency: "never",

                //lastModified: // TODO: ?????
                priority: 0.1
            }
        }),
        // Public static pages and documents
        ...publicFiles.map(path => {
            return {
                url: `${BaseExternalUrl}/${path}`,
                changeFrequency: "never",
                //lastModified: // TODO: ?????
                priority: 0.1
            }
        })
    ]
}