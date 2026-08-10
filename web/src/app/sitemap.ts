import type { MetadataRoute } from 'next'
import {BaseExternalUrl} from "@/app/components/Constants";

export async function generateSitemaps() {
    // Fetch the total number of products and calculate the number of sitemaps needed
    return [{ id: 0 }, { id: 1 }, { id: 2 }, { id: 3 }]
}

export default async function sitemap(props: {
    id: Promise<string>
}): Promise<MetadataRoute.Sitemap> {
    const id = await props.id
    // Google's limit is 50,000 URLs per sitemap
    //const start = id * 50000
    //const end = start + 50000
    // const products = await getProducts(
    //     `SELECT id, date FROM products WHERE id BETWEEN ${start} AND ${end}`
    // )
    const listTyps = ["agarBatches","agarRecipes","bags","fruits","fruitingChambers","grainBatches","jars","jarRecipes","lcs","lcRecipes","lcSyringes","mss","pcRuns","plates","plugs","projects","sales","slants","species","sporePrints","sporeSwabs","stasisTubes","subspecies","substrateBatches","substrateRecipes","transfers","users","waterJars"]
    const importTyps = ["bag","fruit","fruitingChamber","jar","lc","lcSyringe","mss","plate","plugs","slant","sporePrint","sporeSwab","stasisTube","waterJar"]
    const newTyps = ["agarRecipe","jarRecipe","lcRecipe","pcRun", "plugs", "project", "species", "subspecies", "substrateRecipe", "waterJar"]
    return [
        { // Main page
            url: BaseExternalUrl||"mush.appli.ng",
            //lastModified: // TODO: ?????
            //priority: // TODO:?????
            // alternates?: { // TODO: ???
            //     languages?: Languages<string> | undefined
            // } | undefined,
            // images?: string[] | undefined,// TODO: ???
            // videos?: Videos[] | undefined,// TODO: ???
        },
        { // Login page
            url: BaseExternalUrl + "/login",
            //lastModified: // TODO: ?????
            //priority: // TODO:?????
            // alternates?: { // TODO: ???
            //     languages?: Languages<string> | undefined
            // } | undefined,
            // images?: string[] | undefined,// TODO: ???
            // videos?: Videos[] | undefined,// TODO: ???
        },
        // List pages
        ...listTyps.map(typ=>{
            return {
            url: BaseExternalUrl + "/list/" + typ,
                //lastModified: // TODO: ?????
                //priority: // TODO:?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
                // images?: string[] | undefined,// TODO: ???
                // videos?: Videos[] | undefined,// TODO: ???
            }
        }),
        // Import pages
        ...importTyps.map(typ=>{
            return {
                url: BaseExternalUrl + "/import/" + typ,
                //lastModified: // TODO: ?????
                //priority: // TODO:?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
            }
        }),
        // Creation pages
        ...newTyps.map(typ=>{
            return {
                url: BaseExternalUrl + "/new/" + typ,
                //lastModified: // TODO: ?????
                //priority: // TODO:?????
                // alternates?: { // TODO: ???
                //     languages?: Languages<string> | undefined
                // } | undefined,
            }
        })
        // TODO: view pages?
    ]
}