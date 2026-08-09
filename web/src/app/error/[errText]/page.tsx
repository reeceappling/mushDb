import {Metadata} from "next";
import {mushDbTitle} from "@/app/components/Constants";

export const metadata: Metadata = {
    title: mushDbTitle+" error", // TODO: reverse?
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