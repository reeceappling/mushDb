'use server'

// export async function GoogleAnalyticsId(){
//     return process.env.NEXT_PRIVATE_GOOGLE_ANALYTICS_ID || "none";
// }

export async function BaseInternalUrl() {
    return process.env.NEXT_PRIVATE_BASE_API_URL || "localhost"
}