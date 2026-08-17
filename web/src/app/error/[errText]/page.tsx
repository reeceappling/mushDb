import {Metadata} from "next";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

export const metadata: Metadata = {
    title: "Error",
    description: "An error has occurred" // TODO: make this dynamically using errText?
};
export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        errText: string
    }>
}) {
    const props = await params
    return <h1>{"ERROR: "+props.errText}</h1>
}