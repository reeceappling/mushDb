export const BaseInternalUrl= process.env.NEXT_PRIVATE_BASE_API_URL || "localhost" // TODO: ENSURE THE FALLBACK IS OK
export const BaseExternalUrl= process.env.NEXT_PUBLIC_BASE_API_URL // TODO: CHANGE ME!
export const GoogleApiClient= process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "badClientId"
export const TopPageHeaderLevel = 1 // TODO: probably unnecessary
const ChannelName = "Mushrooms" // TODO: probably unnecessary

// TODO: probably unnecessary
export default function CreateChannel(): BroadcastChannel {
    return new BroadcastChannel(ChannelName);
}