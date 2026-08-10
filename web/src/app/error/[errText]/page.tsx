import {Metadata} from "next";

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
}) { // TODO: DELETEME?
    const props = await params
    return <h1>{"ERROR: "+props.errText}</h1>
}