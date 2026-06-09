'use client'

import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import React, {JSX, useContext, useState} from "react";
import {
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import {PcRunData} from "@/app/components/pcRunServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    dataFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType, RequiredKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {
    AclDisplay,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewWaterJarForm} from "@/app/components/waterJarClient";
import {NewBagForm} from "@/app/components/bagClient";
import {NewJarForm} from "@/app/components/jarClient";
import {NewLcForm} from "@/app/components/lcClient";
import {NewSlantForm} from "@/app/components/slantClient";
import {NewStasisTubeForm} from "@/app/components/stasisTubeClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {BagData} from "@/app/components/bagServer";
import {LcData} from "@/app/components/lcServer";
import {SlantData} from "@/app/components/slantServer";
import {WaterJarData} from "@/app/components/waterJarServer";
import {JarData} from "@/app/components/jarServer";
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertPcRun(input: any): asserts input is PcRunData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['runtimeMinutes', 'number'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('PcRun assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('PcRun assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function PcRunDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput<PcRunData>) {
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFor(data.notes || []), new: []})
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: PcRunData) => {
            setInitial(updated)
            setNotes({existing: dataFor(updated.notes || []), new: []})
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const pcRunUpdate = () => {
            const body: any = {
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("pcRun",initial._id, body, AssertPcRun, allCookies(cookies))
                .then(v=>{
                    updateInitial(new PcRunData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const createdLinkFor = (linkText: string, linkId: string, typ: string) => {
            return <EntryLinkForId props={{openInNewTab:false/* TODO: ok?*/,displayId: linkText, linkId: linkId, entryType: typ}}/>
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
                needsTesting: true,
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
                needsTesting: true,
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
                needsTesting: true,
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
                needsTesting: true,
            },
            // TODO: AREA TO ADD ITEMS TO PC? Such as Slants, AgarBatches?
            //
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
                <ID props={{id:data._id, txt:"PC Run", entryType:"pcRun"}}/>
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
                    <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    pcRunUpdate()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewPcRunForm(
    {handlers}: { handlers: NewEntryInput<PcRunData> }) {
    const [date, setDate] = useState(Date.now())
    const [runTime, setRunTime] = useState("")
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const newPcRunSubmit = (e: React.FormEvent) => {
        e.preventDefault()

        const body = {
            //creationDate: date, // Handled serverside
            runTimeMinutes: runTime,
            notes: notes,
        }
        DoCreateRequest("pcRun", body, AssertPcRun, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return (
        <NewEntryFormWrapper entryType={"pcRun"}>
            <div>{"Creating a new PC Run"}</div>
            <ErrorDisplay err={err}/>
            <DateArea pre={"Date : "} when={date} readonly={false} updateParent={setDate}/>
            {/* RunTime TODO: RETHINK THIS???*/}
            <div>
                <div>{"RunTime (minutes):"}</div>
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
    const linkArea: JSX.Element = <div>{(binaryId !== undefined) ?
        <EntryLinkForId props={{linkId: binaryId, entryType: "pcRun",openInNewTab:false}}/>:
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
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new PcRunData(v)}}/>
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
