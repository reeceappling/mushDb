import {Metadata} from "next";
import {BaseExternalUrl} from "@/app/components/Constants";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

export const metadata: Metadata = {
    title: "Error",
    description: "An error has occurred",
    alternates: {
        canonical: `${BaseExternalUrl}/error/An%20Error%20Has%20Occurred`,
        languages: {
            "en-US": `${BaseExternalUrl}/error/An%20Error%20Has%20Occurred`,
        }
    },
    robots: {
        index: false,
        follow: false,
        nocache: true, // TODO: or false?
    },
};
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        errText: string
    }>
}) {
    // TODO: make it look like a normal page???
    const props = await params
    return <h1>{"ERROR: "+props.errText}</h1>// TODO: if logged in allow them to interact with the rfid stuff?
}