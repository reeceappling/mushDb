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
                                       id, entryType
                                   }: {
                                       id: string;
                                       entryType: string
                                   }
) {
    return <a onClick={(e)=>{
        e.preventDefault(); // TODO: ok?
        e.stopPropagation(); // TODO: ok?
    }} href={BaseExternalUrl + "/view/" + entryType + "/" + id}>{id}</a>
}