import {useContext} from "react";
import {DepthContext} from "@/app/components/formSubcomponents/depthContext/depth";
import {OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";

export default function ID({txt, id, entryType, linkPage, allowOpenMainPage}: {
    id: string;
    txt?: string
    entryType: string
    linkPage?: boolean
    allowOpenMainPage?: boolean
}) {
    const depth = useContext(DepthContext);
    const isTopLevel = (depth <= 1)// TODO: ensure ok (DOES NOT DO WHAT WE WANT FOR LIST PAGES)
    return <div className={"idComponent " + (isTopLevel ? "topLevelId" : "nonTopLevelId")}>
        <div className={"idTxt"}>
            {txt ? txt + ": " : ""}
            {(linkPage && !isTopLevel) ? id : <IdPageLink id={id} entryType={entryType}/>}
        </div>
        {(allowOpenMainPage && !isTopLevel) && <OpenMainPage type={"lc"} linkId={id} redirect={false}/>/* TODO: redirect false ok?*/}
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
    const url = BaseExternalUrl + "/view/" + entryType + "/" + id
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