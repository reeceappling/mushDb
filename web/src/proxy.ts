import {NextRequest, NextResponse} from "next/server";
function withHttp(domain: string):string[]{
    return [domain,"http://"+domain]
}
function withPortEntry(inp: string, port?: number):string[]{
    if (port===undefined){
        return [inp]
    }
    return [inp, inp+":"+port]
}
const externalDomain = process.env.NEXT_PUBLIC_DOMAIN || 'failedToGetExternalDomainInMiddleware';
const apiPort = parseInt(process.env.NEXT_PUBLIC_PORT || "443")
const corsOrigins =
    [
        // Our origins

        ...withPortEntry(externalDomain,apiPort),
        'localhost', // TODO: ensure ok (probably delete)
        ...["web"].flatMap(withHttp).flatMap(v=>withPortEntry(v,3000)), // does not have https, not trailing slashes
        ...["api"].flatMap(withHttp).flatMap(v=>withPortEntry(v,8080)), // does not have https
        // Google login origins
        'google.com',
        'accounts.google.com',
    ]
    //: ['*'];
export function proxy(req: NextRequest) {
    const headers = new Headers(req.headers);
    // Handle internal requests from webserver actions
    if (req.headers.get("x-forwarded-host") === 'web:3000') {
        headers.set('origin', `web:3000`);
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
    // Handle windows that have cors origin stuff
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
