'use client'

import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {JSX, useContext, useState} from "react";
import {
    AddCreatedQuadColFunction,
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import {PcRunData} from "@/app/components/pcRunServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {
    createApiUrlFor,
    dataFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey, updateApiUrlFor
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewWaterJarForm} from "@/app/components/waterJarClient";
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
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {AssertLcSyringe} from "@/app/components/lcSyringeClient";
import {AssertMss} from "@/app/components/mssClient";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

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

export default function PcRunDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {
        AssertPcRun(data)
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFor(data.notes || []), new: []})
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: PcRunData) => {
            setInitial(updated)
            setNotes({existing: dataFor(updated.notes || []), new: []})
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const pcRunUpdate = () => {
            const body: any = {
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("pcRun",initial._id, body, AssertPcRun, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("pcRun",data._id), {
            //     method: 'Post',
            //     body: JSON.stringify({
            //         notes: notes,
            //         acl: MarshalAcl(acl),
            //     }),
            //     headers: clientPostRequestHeaders,
            // })
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertPcRun(entry)
            //         updateInitial(entry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const createdLinkFor = (linkText: string, linkId: string, typ: string) => {
            return <EntryLink props={{displayedId: linkText, linkId: linkId, entryType: typ}}>
                <div>{linkText}</div>
            </EntryLink>
        }
        const onViewCreators: OnViewCreatorTriCol[] = [
            {
                txt: "Create Agar Batch",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewAgarBatchForm pcRunInp={data} handlers={{
                        onCreate: (newBatch: AgarBatchData) => {
                            return onCreate([{
                                typeText: "Agar Batch",
                                node: createdLinkFor(newBatch._id, newBatch._id, "agarBatch")
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            {
                txt: "Create Bag",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewBagForm pcRunIn={data} handlers={{
                        onCreate: (newItem: BagData) => {
                            return onCreate([{typeText: "Bag", node: createdLinkFor(newItem._id, newItem._id, "bag")}], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            {
                txt: "Create Grain Jar",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewJarForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: JarData) => {
                            onCreate([{typeText: "Grain Jar", node: createdLinkFor(newEntry._id, newEntry._id, "jar")}],
                                false)
                        },
                        isTopLevel: false,
                    }}
                    />
                },
            },
            {
                txt: "Create Liquid Culture",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewLcForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: LcData) => {
                            onCreate([{
                                typeText: "Liquid Culture",
                                node: createdLinkFor(newEntry._id, newEntry._id, "lc")
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            // TODO: {txt: "Create Plugs Jar", newCreationArea: (onCreate: AddCreatedFunction) => {
            //         return <NewPlugsForm pcRunInput={data._id} redirectOnCreate={false} onCreate={(newEntry: LcData)=>{
            //             onCreate([{typeText: "Plugs Jar", node: mainPageLinkFor(newEntry._id, "plugs")}]) // TODO: ensure typname is ok
            //         }}/>
            //     }},
            {
                txt: "Create Slant",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewSlantForm handlers={{
                        onCreate: (newEntry: SlantData) => {
                            onCreate([{typeText: "Slant", node: createdLinkFor(newEntry._id, newEntry._id, "slant")}], false)
                        },
                        isTopLevel: false
                    }}/>
                },
            },
            {
                txt: "Create Stasis Tube",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewStasisTubeForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: SlantData) => {
                            onCreate([{
                                typeText: "Stasis Tube",
                                node: createdLinkFor(newEntry._id, newEntry._id, "stasisTube")
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            {
                txt: "Create Water Jar",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewWaterJarForm pcRunIn={data} handlers={{
                        onCreate: (newEntry: WaterJarData) => {
                            onCreate([{
                                typeText: "Water Jar",
                                node: createdLinkFor(newEntry._id, newEntry._id, "waterJar")
                            }], false)
                        },
                        isTopLevel: false
                    }}/>
                },
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
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    pcRunUpdate()
                }}>{"Update"}</button>}
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
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const newPcRunSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        let body = {
            creationDate: date,
            runTime: runTime,
            notes: notes,
        }
        DoCreateRequest("pcRun", body, AssertPcRun, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("pcRun"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body),
        // })
        //     .then(HandleJsonResponse)
        //     .then((newItem) => {
        //         AssertPcRun(newItem)
        //         handlers.onCreate && handlers.onCreate(newItem)
        //     })
        //     .catch(setErr);
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

export function PcRunArea({binaryId, headerLevel, offset}: {
    binaryId?: string,
    headerLevel?: number
    offset?: number
}) {
    let linkArea: JSX.Element = <div>{(binaryId !== undefined) ?
        <EntryLink props={{displayedId: binaryId, linkId: binaryId, entryType: "pcRun"}}>{binaryId}</EntryLink> :
        "unknown"}
    </div>
    return <div className={"pcRunArea"}>
        <div>{"Pc Run ID: "}</div>
        <div>{linkArea}</div>
    </div>
}

export function PcRunListPageTable({data, onClick, withLink}: ListPageItems<PcRunData>) {
    let cols: ListTableColumn<PcRunData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Date", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Runtime Mins", (v) => v.runtimeMinutes),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: PcRunData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "pcRun", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function PcRunSelectorTable({data, onClick}: ListPageItems<PcRunData>) {
    return <PcRunListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function PcRunSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: PcRunData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: PcRunData[]): JSX.Element => {
        return <PcRunSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"pcRun"} entryTypes={"pcRuns"} doSelect={doSelect} asserter={AssertPcRun}
                                   table={table}>
        {allowCreate && <NewPcRunForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
