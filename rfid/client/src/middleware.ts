// This function can be marked `async` if using `await` inside
import {NextRequest, NextResponse} from "next/server";
const corsOrigins =
    [
        'appli.ng',
        'appli.ng',
        'mush.appli.ng',
        'https://mush.appli.ng',
        'http://mush.appli.ng',
        'https://mush.appli.ng/',
        'http://mush.appli.ng/',
        'google.com',
        'accounts.google.com',
        'localhost',
        'web',
        'web:3000',
        'http://web',
        'http://web:3000',
        'https://web:3000',
        'api',
        // 'api:8080',
        'http://api',
        // 'http://api:8080',
        // 'https://api:8080',
    ]
    //: ['*'];
export function middleware(req: NextRequest) {
    const headers = new Headers(req.headers);
    if (req.headers.get("x-forwarded-host") === 'web:3000') {
        headers.set('origin', `web:3000`); // TODO: https???
        return NextResponse.next({ request: { headers } });
    }
    const res = NextResponse.next();
    // Set CORS headers
    res.headers.set('Access-Control-Allow-Credentials', 'true');
    res.headers.set('Access-Control-Allow-Methods', 'GET,DELETE,PATCH,POST,PUT');
    res.headers.set(
        'Access-Control-Allow-Headers',
        'X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version'
    );

    if (typeof window !== "undefined") {
        if (corsOrigins.includes('*') || corsOrigins.includes(origin || '')) {
            res.headers.set('Access-Control-Allow-Origin', origin || '*');
        }
    } else {
        // Anything that may happen non-client-side? // TODO: ensure ok
    }


    return res;
}

export const config = {
    matcher: ['/api/:path*', '/:path*', '/auth/:path*'],
};

// export async function middleware(request: NextRequest) {
//     // TODO: if not authenticated, redirect to auth page!
//     let pathParts = request.nextUrl.pathname.split("/")
//     if (request.nextUrl.pathname.startsWith('/view/')) {
//         if(pathParts.length<3 || !["fruit", "mss", "sporePrint", "agarBatch", "agarRecipe", "bag", "fruitingChamber", "jar", "jarRecipe", "lc", "lcRecipe", "pcRun", "plate", "project", "sale", "slant", "species", "stasisTube", "subspecies", "substrateRecipe"]
//             .includes(pathParts[pathParts.length-2])){
//             return NextResponse.redirect(new URL('/error/invalidItemType-' + pathParts[pathParts.length-2], request.url))
//         }
//         return NextResponse.next() // TODO: DOES THIS WORK?
//     }
//
//     let tempItemType = pathParts[pathParts.length - 1]
//     if (request.nextUrl.pathname.startsWith('/new/')) {
//         // Fruit is not available for new. Only made in bag/box pages
//         // MSS only built-in to sporePrint page?
//         // SporePrint only from fruit page
//         let isSubspeciesWithSpecies = pathParts[pathParts.length - 2] === "subspecies"
//         if (!isSubspeciesWithSpecies && !["agarBatch", "agarRecipe", "bag", "fruitingChamber", "jar", "jarRecipe", "lc", "lcRecipe", "pcRun", "plate", "project", "sale", "slant", "species", "stasisTube", "subspecies", "substrateRecipe"]
//             .includes(tempItemType)) {
//             return NextResponse.redirect(new URL('/error/invalidNewItemType-' + tempItemType, request.url))
//         }
//         return NextResponse.next() // TODO: DOES THIS WORK?
//     }
//
//     if (request.nextUrl.pathname.startsWith('/import/')) {
//         let tempItemType = pathParts[pathParts.length - 1]
//         if (!["bag", "fruit", "fruitingChamber", "jar", "lc", "mss", "plate", "slant", "sporePrint", "stasisTube"]
//             .includes(tempItemType)) {
//             return NextResponse.redirect(new URL('/error/invalidImportItemType-' + tempItemType, request.url))
//         }
//         return NextResponse.next() // TODO: DOES THIS WORK?
//     }
//     return NextResponse.next() // TODO: DOES THIS WORK?
// }
//
// // See "Matching Paths" below to learn more
// export const config = {
//     matcher: '/view/:path*'
// }