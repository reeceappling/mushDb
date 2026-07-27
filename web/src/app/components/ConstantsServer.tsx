'use server'

export async function BaseInternalUrl() {
    return process.env.NEXT_PRIVATE_BASE_API_URL || "localhost"
}