import {ReactNode, useContext} from "react";
import {DepthContext} from "@/app/components/formSubcomponents/depthContext/depth";
import {OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {IsMainCollEntryType, viewUrlFor} from "@/app/components/common";

// ID component should only be used on view pages for the currently-being-viewed ID
export default function ViewPageEntryID({props, children}: {
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
            <div className={"mr-2"}>{props.txt ? <span itemProp={"additionalType"/* TODO: ensure ok*/}>props.txt</span> + ": " : ""}</div>
            <div>{children}</div>
            <div className={"ml-2"}>{children!==undefined&&"("}{(props.linkPage && !isTopLevel) ? /*TODO: microdata here or no?*/props.id : <IdPageLink id={props.id} entryType={props.entryType}/* isMainCollItem={IsMainCollEntryType(props.entryType)}*//>}{children!==undefined&&")"}</div>
        </div>
        {(props.allowOpenMainPage && !isTopLevel) && <OpenMainPage type={props.entryType} linkId={props.id} redirect={false}/>/* TODO: redirect false ok?*/}
    </div>
}

export function IdPageLink({
                                       id, entryType, openInNewTab, itemProp, itemScope, itemType
                                   }: {
                                       id: string;
                                       entryType: string;
                                       //isMainCollItem: boolean;
                                       openInNewTab?: boolean
                                       itemProp?: string
                                       itemScope?: boolean
                                       itemType?: string
                                   }
) {
    const url = viewUrlFor(entryType, id)
    const onClickStopPropagation = (e: React.MouseEvent)=>{
        e.preventDefault(); // TODO: is this ok?
        e.stopPropagation();
    }
    //const itemType = "" // TODO: SET THIS! https://mush.appli.ng/schemas/${entryType}

    // TODO: validate both work!
    if (openInNewTab){
        return <a itemScope={itemScope} itemType={itemType} itemProp={itemProp} itemID={url} href={url} target={"_blank"} rel={"noopener noreferrer"} onClick={onClickStopPropagation}>{id}</a>
    }
    return <a itemScope={itemScope} itemType={itemType} itemProp={itemProp} itemID={url} href={url} onClick={onClickStopPropagation}>{id}</a>// TODO: ALTERNATE ID?
}