export const BaseExternalDomain= process.env.NEXT_PUBLIC_DOMAIN || "localhost" // TODO: USE?
export const BaseExternalPort= process.env.NEXT_PUBLIC_PORT || "443"
export const BaseInternalPort= 8080
export const BaseInternalUrl= process.env.NEXT_PRIVATE_BASE_API_URL || "localhost"
export const BaseExternalUrl= process.env.NEXT_PUBLIC_BASE_API_URL
export const GoogleApiClient= process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "badClientId"
export const TopPageHeaderLevel = 1 // TODO: probably unnecessary
const ChannelName = "Mushrooms" // TODO: probably unnecessary

// TODO: probably unnecessary
export default function CreateChannel(): BroadcastChannel {
    return new BroadcastChannel(ChannelName);
}