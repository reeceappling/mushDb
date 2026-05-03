'use client'

import NotesAreaOld, {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {JSX, useState} from "react";
import {
    AddCreatedTriColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import {PcRunData} from "@/app/components/pcRunServer";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey
} from "@/app/components/common";
import {ErrorDisplay, NameArea} from "@/app/components/formSubcomponents/commonClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewWaterJarForm} from "@/app/components/waterJarClient";
import {FlexedArea, FlexedSinglesGroup, NewAgarBatchForm, NotesFormArea} from "@/app/components/agarBatchClient";
import {NewBagForm} from "@/app/components/bagClient";
import {NewJarForm} from "@/app/components/jarClient";
import {NewLcForm} from "@/app/components/lcClient";
import {NewSlantForm} from "@/app/components/slantClient";
import {NewStasisTubeForm} from "@/app/components/stasisTubeClient";
import TestAndValidate from "@/app/components/testing/untested";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {BagData} from "@/app/components/bagServer";
import {LcData} from "@/app/components/lcServer";
import {SlantData} from "@/app/components/slantServer";
import {WaterJarData} from "@/app/components/waterJarServer";
import {JarData} from "@/app/components/jarServer";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {DisplayFormWrapper, NewEntryFormWrapper} from "./lcRecipeClient";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {MssData} from "@/app/components/mssServer";

export function AssertPcRun(input: any): asserts input is PcRunData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['runtimeMinutes', 'number'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('PcRun assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('PcRun assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
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
    return <div>{"To " + dstType + " "}<a href={BaseExternalUrl + "/view/" + dstType + "/" + id}>{id}</a></div>
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
    return <DepthProvider>
        <div className={"ovcArea"}>{/* TODO: style ovcArea*/}
            {props.children}
        </div>
    </DepthProvider>
}

// TODO: MOVE!
/* View lc/2Aui6ejTFsd for testing */
export function OnViewCreatorsQuadColArea({OnViewCreators,readonly}: { OnViewCreators: OnViewCreatorQuadCol[], readonly: boolean}) {
    if (readonly) {
        return null
    }
    const [activeTab, setActiveTab] = useState<string | undefined>();
    const [created, setCreated] = useState<CreatedLinkQuadCol[]>([]);
    const [createdTableHidden, setCreatedTableHidden] = useState<boolean>(false);
    const addCreated = (newLinks: CreatedLinkQuadCol[]) => {
        setCreated(created.concat(newLinks))
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
        <DepthProvider><OvcArea> {/* TODO: style ovcArea*/}
            <OvcTopBar setActiveTab={setActiveTab} OnViewCreators={OnViewCreators} hasExtraCol={true}
                       activeTab={activeTab}/>
            <OvcLinksTableQuad created={created} tableHidden={createdTableHidden} toggleHidden={toggleHidden}/>
            {creatorBody()}
        </OvcArea>
        </DepthProvider>
    </TestAndValidate>
}

// TODO: MOVE!
export function OnViewCreatorsTriColArea({OnViewCreators, readonly}: { OnViewCreators: OnViewCreatorTriCol[], readonly: boolean }) {
    if (readonly) {
        return null
    }
    const [activeTab, setActiveTab] = useState<string | undefined>();
    const [created, setCreated] = useState<CreatedLinkTriCol[]>([]);
    const [createdTableHidden, setCreatedTableHidden] = useState<boolean>(false);
    const addCreated = (newLinks: CreatedLinkTriCol[]) => {
        setCreated(created.concat(newLinks))
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
        return <OvcCreatorBodyWrapper> {/* TODO: style ovcCreatorBodyWrapper*/}
            {closeButton}
            {creator.newCreationArea(addCreated)}
            {closeButton}
        </OvcCreatorBodyWrapper>
    }
    return <TestAndValidate todos={["TEST THIS WHOLE THING!"]}><OvcArea>

        <OvcTopBar setActiveTab={setActiveTab} OnViewCreators={OnViewCreators} hasExtraCol={false}
                   activeTab={activeTab}/>
        <OvcLinksTableTri created={created} tableHidden={createdTableHidden} toggleHidden={toggleHidden}/>
        {creatorBody()}

    </OvcArea></TestAndValidate>
}

// TODO: MOVE
function OnViewCreatorCloseButton({handleClose, activeTab}: { handleClose: () => void, activeTab?: string }) {
    return <button className={"basicButton"} onClick={handleClose}>
        {'Close ' + (activeTab !== undefined ? ('"' + activeTab + '" ') : "") + " Area"}
    </button>
}

// TODO: MOVE
export function HiddenDiv() {
    return <div className={"hidden"}></div>
}

// TODO: MOVE
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

export default function PcRunDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {
        AssertPcRun(data)
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>({existing:dataFor(data.notes || []),new:[]})
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial= (updated: PcRunData)=>{
            setInitial(updated)
            setNotes({existing:dataFor(updated.notes || []),new:[]})
            setAcl(updated.acl)
        }
        const pcRunUpdate = () => {
            // TODO: COMPARE NEW NOTES?
            fetch(BaseExternalUrl + "/db/update/pcRun/" + data._id, {
                method: 'Post',
                body: JSON.stringify({ // TODO: used to just be notes! Fix in go side if needed
                    notes: notes,
                    acl: MarshalAcl(acl),
                }),
                headers: {
                    credentials: 'include',
                    // TODO: may need 'Cookie': cookies,
                    'Content-type': "application/json"
                },
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertPcRun(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const createdLinkFor = (linkText: string, linkId: string, typ: string) => {
            return <EntryLink props={{displayedId: linkText, linkId: linkId, entryType: typ}}>
                <div>{linkText}</div>
            </EntryLink>
        }
        const onViewCreators: OnViewCreatorTriCol[] = [
            {
                txt: "Create Agar Batch", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewAgarBatchForm pcRunInp={data} handlers={{
                        onCreate: (newBatch: AgarBatchData) => {
                            return onCreate([{
                                typeText: "Agar Batch",
                                node: createdLinkFor(newBatch._id, newBatch._id, "agarBatch")
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            },
            {
                txt: "Create Bag", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewBagForm pcRunIn={data} handlers={{
                        onCreate: (newItem: BagData) => {
                            return onCreate([{typeText: "Bag", node: createdLinkFor(newItem._id, newItem._id, "bag")}])
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            {
                txt: "Create Grain Jar", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewJarForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: JarData) => {
                            onCreate([{typeText: "Grain Jar", node: createdLinkFor(newEntry._id, newEntry._id, "jar")}])
                        },
                        isTopLevel: false,
                    }}
                    />
                }
            },
            {
                txt: "Create Liquid Culture", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewLcForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: LcData) => {
                            onCreate([{
                                typeText: "Liquid Culture",
                                node: createdLinkFor(newEntry._id, newEntry._id, "lc")
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            },
            // TODO: {txt: "Create Plugs Jar", newCreationArea: (onCreate: AddCreatedFunction) => {
            //         return <NewPlugsForm pcRunInput={data._id} redirectOnCreate={false} onCreate={(newEntry: LcData)=>{
            //             onCreate([{typeText: "Plugs Jar", node: mainPageLinkFor(newEntry._id, "plugs")}]) // TODO: ensure typname is ok
            //         }}/>
            //     }},
            {
                txt: "Create Slant", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewSlantForm handlers={{
                        onCreate: (newEntry: SlantData) => {
                            onCreate([{typeText: "Slant", node: createdLinkFor(newEntry._id, newEntry._id, "slant")}])
                        },
                        isTopLevel: false
                    }}/>
                }
            },
            {
                txt: "Create Stasis Tube", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewStasisTubeForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: SlantData) => { // TODO: allow pcRun input?
                            onCreate([{
                                typeText: "Stasis Tube",
                                node: createdLinkFor(newEntry._id, newEntry._id, "stasisTube")
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            },
            {
                txt: "Create Water Jar", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewWaterJarForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: WaterJarData) => { // TODO: allow pcRun input?
                            onCreate([{
                                typeText: "Water Jar",
                                node: createdLinkFor(newEntry._id, newEntry._id, "waterJar")
                            }])
                        },
                        isTopLevel: false
                    }}/>
                }
            },
        ]
        return (
            <DisplayFormWrapper entryType={"pcRun"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"PC Run"} entryType={"pcRun"}/>
                <OnViewCreatorsTriColArea OnViewCreators={onViewCreators} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Run Date: "} when={initial.creationDate} readonly={true}/>
                        <div>{"Run Time: " + initial.runtimeMinutes}</div>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <input type="submit" value="Update" onClick={pcRunUpdate} onSubmit={(e) => {
                    e.preventDefault();
                }}/>}
            </DisplayFormWrapper>
        )
    } catch
        (err) {
        return <div>{"ERROR: PC Run data format incorrect: " + err}</div>
    }
}

export function NewPcRunForm(
    {handlers}: { handlers: NewEntryInput<PcRunData> }) {
    const [date, setDate] = useState(Date.now())
    const [runTime, setRunTime] = useState("")
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    // TODO: handle isTopLevel
    const newPcRunSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        let body = {creationDate: date, runTime: runTime, notes: notes}
        fetch(BaseExternalUrl + "/create/pcRun", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body),
        })
            .then(HandleJsonResponse)
            .then((newItem) => {
                AssertPcRun(newItem)
                handlers.onCreate && handlers.onCreate(newItem)
            })
            .catch(setErr);
    }
    return (
        <NewEntryFormWrapper entryType={"pcRun"}>
            <div>{"Creating a new PC Run"}</div>
            <ErrorDisplay err={err}/>
            <DateArea pre={"Date : "} when={date} readonly={false} updateParent={setDate}/>
            {/* RunTime TODO: RETHINK THIS???*/}
            <div>
                <div>{"RunTime:"}</div>
                {/* TODO: Consider runtimeMinutes? ?*/}
                <input type="text" value={runTime} onChange={(e) => {
                    setRunTime(e.currentTarget.value)
                }}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" onClick={newPcRunSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function PcRunInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<PcRunData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"PC Run"} entryType={"pcRun"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <DateArea pre={"Date: "} when={data.creationDate} readonly={true}/>
            <NameArea headerTxt={"Runtime: "} currentName={data.runtimeMinutes.toString()} readonly={true}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                                                     expanded={expanded}/>
    </InlineEntry>
}

export function PcRunArea({binaryId, headerLevel, offset}: {
    binaryId?: string,
    headerLevel?: number
    offset?: number
}) {
    // TODO: does this need depth?
    let linkArea: JSX.Element = <div>{(binaryId !== undefined)?
        <EntryLink props={{displayedId: binaryId, linkId: binaryId, entryType: "pcRun"}}>{binaryId}</EntryLink>:
        "unknown"}
    </div>
    return <div className={"pcRunArea"}>
        <div>{"Pc Run ID: "}</div>
        <div>{linkArea}</div>
    </div>
}

// TODO: PC run selector????

// export function PcRunListDisplay(ps: SingleListProps<PcRunData>) {
//     return <div>
//         {ps.data.map((b, i) => {
//             return <PcRunInline data={b} onClick={() => {
//                 ps.onClick(b)
//             }} key={i}/>
//         })}
//     </div>
// }
