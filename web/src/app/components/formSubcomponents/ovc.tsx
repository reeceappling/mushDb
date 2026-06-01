'use client'

import {
    AddCreatedQuadColFunction,
    OnViewCreatorQuadCol,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import {NewFruitForm} from "@/app/components/fruitClient";
import {FruitData} from "@/app/components/fruitServer";
import {TransferData} from "@/app/components/transferServer";
import {NewTransferArea} from "@/app/components/transferClient";
import {CreatedLinkFor, Subform, viewUrlFor} from "@/app/components/common";
import {JSX, useState} from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
import TestAndValidate from "@/app/components/testing/untested";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";

export function OvcForXfers(parentId: string, parentType: string, validTypesTo: string[], cookies: string, addTransferOut?: (xfer: TransferData) => void, altTxt?: string): OnViewCreatorQuadCol {
    return {
        txt: altTxt || "New Transfer",
        newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
            return <NewTransferArea cookies={cookies} idFrom={parentId} typeFrom={parentType}
                                    validTypesTo={validTypesTo}
                                    onCreated={(xfer: TransferData) => {
                                        addTransferOut && addTransferOut(xfer)
                                        onCreate([{
                                            typeText: "Transfer",
                                            node: <CreatedLinkFor linkId={xfer._id} typ={"transfer"}/>,
                                            lastNode: <QuadColLastCol dstType={xfer.toType} id={xfer.to}/>
                                        }], false)
                                    }}/>
        }
    }
}

// TODO: MOVE!
export function OvcForNewFruit(parentId: string, parentType: string, cookies: string): OnViewCreatorQuadCol {
    return {
        txt: "Record New Fruit",
        newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
            return <NewFruitForm parentId={parentId} parentType={parentType} readonly={false} cookies={cookies}
                                 onCreate={(fr: FruitData) => {
                                     onCreate([{
                                         typeText: "Fruit",
                                         node: <CreatedLinkFor linkId={fr._id} typ={"fruit"}/>,
                                     }], false)
                                 }}/>
        },
    }
}

// TODO: MOVE
export type CreatedLinkTriCol = {
    typeText: string,
    node: JSX.Element,
}
// TODO: MOVE
type CreatedLinkExtraCol = {
    lastNode?: JSX.Element,
}
// TODO: MOVE
export type CreatedLinkQuadCol = CreatedLinkTriCol & CreatedLinkExtraCol

// TODO: MOVE
export function QuadColLastCol({dstType, id}: { dstType: string, id: string }) {
    return <div>{"To " + dstType + " "}<a href={viewUrlFor(dstType, id)}>{id}</a></div>
}

// TODO: MOVE
function OvcQuadRow({item, key}: { item: CreatedLinkQuadCol, key: number }) {
    const emptyCell = "-" // TODO: ensure ok
    return <OvcTriRow item={item} key={key}>
        <td>{item.lastNode || emptyCell}</td>
    </OvcTriRow>
}

// TODO: MOVE
function OvcTriRow(props: React.PropsWithChildren<{ item: CreatedLinkTriCol, key: number }>) {
    return <tr key={props.key}>
        {/* TODO: styling for table data (non-first rows)*/}
        <td>{props.item.typeText}</td>
        <td>{props.item.node}</td>
        {props.children}
    </tr>
}

// TODO: MOVE
function OvcTableHidden({empty, unhide}: { empty: boolean, unhide: () => void }) {
    return <div className={empty ? "hidden" : ""/*hide button if no entries*/} onClick={unhide}>
        {"Show Created Entries Table"}
    </div> // TODO: ensure hidden works
}

// TODO: MOVE
function OvcLinksTableWrapper(props: React.PropsWithChildren<{
    created: CreatedLinkTriCol[],
    hidden: boolean,
    toggleHidden: () => void
}>) {
    if (props.hidden || props.created.length === 0) {
        return <OvcTableHidden empty={props.created.length === 0} unhide={props.toggleHidden}/>
    }
    return <div>
        <div className={"areaHeader"}>{"Entries Created:"}</div>
        {/* TODO: ok?*/}
        <table className={"ovcLinksTable"}>
            {props.children}
        </table>
        {/* TODO: styling for outputs table*/}
    </div>
}

// TODO: MOVE
const OvcHideTableText = "Hide Table"

// TODO: MOVE
function OvcTableHeaders({headersTxt, setTableHidden}: { headersTxt: string[], setTableHidden: () => void }) {
    return <tr>
        {/* TODO: styling for table headers*/}
        {headersTxt.map((txt, i) => {
            if (txt === OvcHideTableText) { //TODO: hide button styling and hover styling
                return <th key={i} onClick={setTableHidden}>{OvcHideTableText}</th>
            }
            return <th key={i}>{txt}</th>
        })}
    </tr>
}

// TODO: MOVE
function OvcLinksTableQuad(
    {created, tableHidden, toggleHidden}: {
        created: CreatedLinkQuadCol[],
        tableHidden: boolean,
        toggleHidden: () => void
    }) {
    return <OvcLinksTableWrapper created={created} hidden={tableHidden} toggleHidden={toggleHidden}>
        <OvcTableHeaders headersTxt={["Created", "Link", OvcHideTableText]} setTableHidden={toggleHidden}/>
        <tbody>
        {created.map((createdEntry, i) => {
            return <OvcQuadRow item={createdEntry} key={i}/>
        })}
        </tbody>
    </OvcLinksTableWrapper>
}

// TODO: MOVE
function OvcLinksTableTri(
    {created, tableHidden, toggleHidden}: {
        created: CreatedLinkTriCol[],
        tableHidden: boolean,
        toggleHidden: () => void
    }) {
    return <OvcLinksTableWrapper created={created} hidden={tableHidden} toggleHidden={toggleHidden}>
        <OvcTableHeaders headersTxt={["Created", OvcHideTableText]} setTableHidden={toggleHidden}/>
        <tbody>
        {created.map((createdEntry, i) => {
            return <OvcTriRow item={createdEntry} key={i}/>
        })}
        </tbody>
    </OvcLinksTableWrapper>
}

// TODO: MOVE
function OvcCreatorBodyWrapper(props: React.PropsWithChildren<{}>) {
    return <div className={"ovcCreatorBodyWrapper"}>{/* TODO: style ovcCreatorBodyWrapper*/}
        {props.children}
    </div>
}

// TODO: MOVE
function OvcArea(props: React.PropsWithChildren<{}>) {
    return <Subform>
        <div className={"ovcArea depth"}>{/* TODO: style ovcArea*/}
            {props.children}
        </div>
    </Subform>
}

// TODO: MOVE!
/* View lc/2Aui6ejTFsd for testing */
export function OnViewCreatorsQuadColArea({OnViewCreators, readonly}: {
    OnViewCreators: OnViewCreatorQuadCol[],
    readonly: boolean
}) {
    if (readonly || !OnViewCreators) {
        return null
    }
    const [activeTab, setActiveTab] = useState<string | undefined>();
    const [created, setCreated] = useState<CreatedLinkQuadCol[]>([]);
    const [createdTableHidden, setCreatedTableHidden] = useState<boolean>(false);
    const addCreated: AddCreatedQuadColFunction = (newLinks: CreatedLinkQuadCol[], closeAfter:boolean) => {
        setCreated(created.concat(newLinks))
        if (closeAfter) {
            setActiveTab(undefined)
        }
    }
    const toggleHidden = () => {
        setCreatedTableHidden(!createdTableHidden)
    }
    const closeButton = <OnViewCreatorCloseButton handleClose={() => {
        setActiveTab(undefined)
    }} activeTab={activeTab}/>

    const creatorBody = () => {
        if (activeTab === undefined) {
            return <HiddenDiv/>
        }
        const creator = OnViewCreators.find(ovc => activeTab === ovc.txt)
        if (creator === undefined) {
            console.error("could not find ovc for " + activeTab + " in tab options")
            return <HiddenDiv/>
        }

        return <OvcCreatorBodyWrapper>
            {closeButton}
            {creator.newCreationArea(addCreated)}
            {closeButton}
        </OvcCreatorBodyWrapper>
    }
    return <TestAndValidate todos={["TEST THIS WHOLE THING!"]}>
        <DepthProvider><OvcArea>
            <OvcTopBar setActiveTab={setActiveTab} OnViewCreators={OnViewCreators} hasExtraCol={true}
                       activeTab={activeTab}/>
            <OvcLinksTableQuad created={created} tableHidden={createdTableHidden} toggleHidden={toggleHidden}/>
            {creatorBody()}
        </OvcArea>
        </DepthProvider>
    </TestAndValidate>
}

export function OnViewCreatorsTriColArea({OnViewCreators, readonly}: {
    OnViewCreators: OnViewCreatorTriCol[],
    readonly: boolean
}) {
    if (readonly || !OnViewCreators) {
        return null
    }
    const [activeTab, setActiveTab] = useState<string | undefined>();
    const [created, setCreated] = useState<CreatedLinkTriCol[]>([]);
    const [createdTableHidden, setCreatedTableHidden] = useState<boolean>(false);
    const addCreated = (newLinks: CreatedLinkTriCol[], closeAfter: boolean) => {
        setCreated(created.concat(newLinks))
        if (closeAfter) {
            setActiveTab(undefined)
        }
    }
    const toggleHidden = () => {
        setCreatedTableHidden(!createdTableHidden)
    }
    // TODO: Top level hidden instead of dynamic DOM to reduce client strain?

    const creatorBody = () => {
        if (activeTab === undefined) {
            return <HiddenDiv/> // Can be null because it is the last subcomponent
        }
        const creator = OnViewCreators.find(ovc => activeTab === ovc.txt)
        if (creator === undefined) {
            console.error("could not find ovc for " + activeTab + " in tab options")
            return <HiddenDiv/>
        }
        const closeButton = <OnViewCreatorCloseButton handleClose={() => {
            setActiveTab(undefined)
        }} activeTab={activeTab}/>
        return <OvcCreatorBodyWrapper>
            {closeButton}
            {creator.newCreationArea(addCreated)}
            {closeButton}
        </OvcCreatorBodyWrapper>
    }
    return <TestAndValidate todos={["TEST THIS WHOLE THING!"]}>
        <OvcArea>
            <OvcTopBar setActiveTab={setActiveTab} OnViewCreators={OnViewCreators} hasExtraCol={false}
                       activeTab={activeTab}/>
            <OvcLinksTableTri created={created} tableHidden={createdTableHidden} toggleHidden={toggleHidden}/>
            {creatorBody()}
        </OvcArea>
    </TestAndValidate>
}

function OnViewCreatorCloseButton({handleClose, activeTab}: { handleClose: () => void, activeTab?: string }) {
    return <button className={"basicButton"} onClick={handleClose}>
        {'Close ' + (activeTab !== undefined ? ('"' + activeTab + '" ') : "") + " Area"}
    </button>
}

export function HiddenDiv() {
    return <div className={"hidden"}></div>
}

function OvcTopBar({activeTab, setActiveTab, OnViewCreators, hasExtraCol}: {
    activeTab?: string,
    setActiveTab: (nat?: string) => void,
    OnViewCreators: OnViewCreatorTriCol[],
    hasExtraCol: boolean
}) {
    return <div className={"ovcBar " + (hasExtraCol ? "ovcBarQuad" : "ovcBarTri")}>
        {/* TODO: styling? OnHover, onClick*/}
        {OnViewCreators.map((ovc, i) => {
            const isActiveTab = (ovc.txt === activeTab)
            const classes = "ovcBarItem " + (isActiveTab ? "currentlyActive" : "selectable") // TODO: ovcBarItem, currentlyActive, selectable
            const onClick = isActiveTab ? () => {
            } : () => {
                setActiveTab(ovc.txt)
            }
            return <div key={ovc.txt} className={classes} onClick={onClick}>{ovc.txt}</div>
        })}
    </div>
}