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