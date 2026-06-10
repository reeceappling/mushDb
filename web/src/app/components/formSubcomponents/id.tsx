import {ReactNode, useContext} from "react";
import {DepthContext} from "@/app/components/formSubcomponents/depthContext/depth";
import {OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {viewUrlFor} from "@/app/components/common";

export default function ID({props, children}: {
    props: {
        id: string;
        txt?: string
        entryType: string
        linkPage?: boolean
        allowOpenMainPage?: boolean
    },
    children?: ReactNode,

}) {
    const depth = useContext(DepthContext);
    const isTopLevel = (depth <= 1)// TODO: ensure ok (DOES NOT DO WHAT WE WANT FOR LIST PAGES)
    return <div className={"idComponent " + (isTopLevel ? "topLevelId" : "nonTopLevelId")}>
        <div className={"idTxt inlineChildren"}>
            <div className={"mr-2"}>{props.txt ? props.txt + ": " : ""}</div>
            <div>{children}</div>
            <div className={"ml-2"}>{children!==undefined&&"("}{(props.linkPage && !isTopLevel) ? props.id : <IdPageLink id={props.id} entryType={props.entryType}/>}{children!==undefined&&")"}</div>
        </div>
        {(props.allowOpenMainPage && !isTopLevel) && <OpenMainPage type={"lc"} linkId={props.id} redirect={false}/>/* TODO: redirect false ok?*/}
    </div>
}

export function IdPageLink({
                                       id, entryType, openInNewTab
                                   }: {
                                       id: string;
                                       entryType: string;
                                       openInNewTab?: boolean // TODO: use this everywhere needed 5/21/26
                                   }
) {
    const url = viewUrlFor(entryType, id)
    const onClickStopPropagation = (e: React.MouseEvent)=>{
        e.preventDefault(); // TODO: ok?
        e.stopPropagation();
    }
    // TODO: validate both work!
    if (openInNewTab){
        return <a href={url} target={"_blank"} rel={"noopener noreferrer"} onClick={onClickStopPropagation}>{id}</a>
    }
    return <a href={url} onClick={onClickStopPropagation}>{id}</a>
}