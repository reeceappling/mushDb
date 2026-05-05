'use client'

import React, {JSX, useEffect, useState} from "react";
import {useQuery,} from '@tanstack/react-query'
import NotesArea, {IsValidNote, Note, NotesAreaInline, NotesAreaOld} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {AgarBatchData, AgarColor} from "@/app/components/agarBatchServer";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
} from "@/app/components/common";
import {AgarRecipeArea, AgarRecipeSelector, dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {OnViewCreatorsTriColArea, PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, RecentPCRunSelector} from "@/app/components/pcRunServer";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SelectorFor} from "@/app/components/selector";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth,} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewPlateForm} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewSlantForm} from "@/app/components/slantClient";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";

export function AssertAgarBatch(input: any): asserts input is AgarBatchData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['color', 'string'],
        ['pcRun', 'string'],
        ['agarRecipe', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function AgarBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertAgarBatch(data)

        const [initial, setInitial] = useState(data)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: AgarBatchData) => {
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const agarBatchSubmit = () => {
            if (notes.new.length === 0 && notes.existing === dataFor(initial.notes)) {
                setErr("No changes found")
                return
            }
            fetch(BaseExternalUrl + "/db/update/agarBatch/" + initial._id, { // This ID is in base58
                method: 'Post',
                body: JSON.stringify({notes: notes, acl: acl}),
                headers: {
                    credentials: 'include',
                    'Content-type': 'application/json'
                },
            })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertAgarBatch(newEntry)
                    updateInitial(newEntry)
                })
                .catch((err) => {
                    setErr(JSON.stringify(err))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "Create Plates", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewPlateForm agarBatchIn={data} handlers={{
                        onCreate: (newItem: PlateData) => {
                            return onCreate([{
                                typeText: "Plate",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"plate"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            },
            {
                // TODO: CreateSlants. Both agar and slants will either be PC-d same time in seperate containers or agar PC'd while inside the slant...
                txt: "Create Slants", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewSlantForm agarBatchIn={data} handlers={{
                        onCreate: (newItem: PlateData) => {
                            return onCreate([{
                                typeText: "Slant",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"slant"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            },
        ]
        return (
            <DisplayFormWrapper entryType={"agarBatch"}>
                <ID txt={"Agar Batch"} id={data._id} entryType={"agarBatch"} linkPage={false} allowOpenMainPage={false}
                    data-cy-id={"Id"}/>
                <ErrorDisplay data-cy-id={"Error"} err={err} headerLevel={headerLevel}/>
                <FlexedArea>
                    <DateArea data-cy-id={"LastUpdated"} pre={"Last Updated: "} when={initial.lastUpdated}
                              readonly={true}/>{/* TODO: ensure this is now inline*/}
                    <div data-cy-id={"Color"}>{"Color: " + data.color}</div>
                    <PcRunArea data-cy-id={"Run"} binaryId={initial.pcRun} headerLevel={headerLevel}/>
                    <AgarRecipeArea data-cy-id={"Recipe"} agarRecipeBinId={initial.agarRecipe}/>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>

                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    agarBatchSubmit()
                }}>{"Update"}</button>}
                {/* TODO: styling for onViewCreators!!!!*/}
                {/* TODO: REFORMAT ALL LIST PAGES!!!!*/}
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: AgarBatch data format incorrect: " + err}</div>
    }
}

export function NotesFormArea({ // TODO: CURRENTLY DOES NOT WORK PROPERLY WHEN SOME NOTES ARE DELETED, FIX!
                                  readonly,
                                  initial,
                                  updateParent,
                              }: {
    readonly?: boolean,
    initial?: Note[], // TODO: ensure everywhere is using this properly
    updateParent?: (entries: AllEntries<Note>) => void,
}) {
    const [count, setCount] = useState(0)
    useEffect(() => {
        setCount(count + 1);
    }, [initial]);
    return <div key={count}>
        <div className={"areaHeader"/* TODO: ok? */}>{"Notes"}</div>
        <NotesArea data-cy-id={"Notes"} readonly={readonly} initial={initial} updateParent={updateParent}/>
    </div>
}

// TODO: MOVE
export function FlexedArea(props: React.PropsWithChildren<{}>) {
    return <div className={"flexedArea"}>{props.children}</div>
}

// TODO: MOVE
export function FlexedSinglesGroup(props: React.PropsWithChildren<{}>) {
    return <div className={"flexedSinglesGroup"}>{props.children}</div>
}

// TODO: NOT WORKING IN SELECTOR!
export function NewAgarBatchForm({handlers, agarRecipeIn, pcRunInp}: {
    handlers: NewEntryInput<AgarBatchData>,
    agarRecipeIn?: AgarRecipeData,
    pcRunInp?: PcRunData
}) {
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunInp)
    const [recipe, setRecipe] = useState<AgarRecipeData | undefined>(agarRecipeIn)
    const [color, setColor] = useState<AgarColor>("Clear")
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    // TODO: handle isTopLevel
    const newAgarBatchSubmit = () => {
        // pcRun, recipe must exist
        if (!pcRun) {
            setErr("pcRun must be selected")
            return
        }
        if (!recipe) {
            setErr("recipe must be selected")
            return
        }
        let body: any = {
            color: color,
            pcRun: pcRun._id,
            recipe: recipe._id,
            notes: notes,
        }
        fetch(BaseExternalUrl + "/create/agarBatch", {
            method: 'Post',
            body: JSON.stringify(body),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            }
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertAgarBatch(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <NewEntryFormWrapper entryType={"agarBatch"}>
        <div data-cy-id="Header">{"Creating a new agar batch"}</div>
        <ErrorDisplay data-cy-id="Error" err={err}/>
        {pcRunInp === undefined ?
            <RecentPCRunSelector data-cy-id="PcRun" doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                 creatorInPage={handlers.isTopLevel}/* TODO: isTopLevel*//>
            : <PcRunArea binaryId={pcRunInp?._id}/>
        }
        {agarRecipeIn === undefined ?
            <AgarRecipeSelector data-cy-id="Recipe" doSelect={setRecipe} showRecent={handlers.isTopLevel}
                                allowCreate={handlers.isTopLevel}/* TODO: isTopLevel*//>
            : <AgarRecipeArea agarRecipeBinId={agarRecipeIn?._id}/>
        }
        <AgarColorArea data-cy-id={"Color"} current={color} onSelect={setColor}/>
        <NotesAreaOld data-cy-id="Notes" readonly={false} updateParent={v => { // TODO: notesFormArea?
            setNotes(v.new.map(x => {
                return x.data
            }))
        }}/>
        <button className={"bottomButton greenButton"} onClick={newAgarBatchSubmit}>{"Submit"}</button>
    </NewEntryFormWrapper>
}

export function AgarBatchInline({
                                    data,
                                    expandByDefault,
                                    onClick,
                                    showMainPageButton,
                                    idIsLink
                                }: InlineProps<AgarBatchData>) {
    // TODO: DO INLINES NEED DEPTH PROVIDERS??????
    const notes = data.notes || []
    const [expanded, setExpanded] = useState(expandByDefault)
    const areaProps = () => {
        return {expanded: expanded, setExpanded: setExpanded}
    }
    const pcRunDisplayId = data.pcRun
    return <InlineEntry onClick={onClick}>
        <InlineSubArea data-cy-id="InlineTop" props={{}}> {/* TODO: do we need data-cy-id on this?*/}
            <ID data-cy-id="Id" id={data._id} txt={"Agar Batch"} entryType={"agarBatch"}
                allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <div data-cy-id="Recipe">
                <EntryLink props={{displayedId: data.agarRecipe, linkId: data.agarRecipe, entryType: "agarRecipe"}}>
                    <div>{data.agarRecipe}</div>
                </EntryLink>
            </div>
            <div data-cy-id="Color">{data.color}</div>
        </InlineSubArea>
        <InlineExpansionArea data-cy-id="InlineBottom" props={areaProps()}> {/* TODO: do we need data-cy-id on this?*/}
            <div data-cy-id="PcRun" className={"inline"}>
                <EntryLink props={{displayedId: pcRunDisplayId, linkId: pcRunDisplayId, entryType: "pcRun"}}>
                    <div>{pcRunDisplayId}</div>
                </EntryLink>
            </div>
            <NotesAreaInline data-cy-id="Notes" notes={notes} offset={-1}/>
            <DateArea data-cy-id="LastUpdated" pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
            <button className={"basicButton"} data-cy-id="CloseButton" onClick={(e) => {
                e.stopPropagation();
                setExpanded(false)
            }}>{"See Less"}</button>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// TODO: MOVE!
export function ListPageTableRow<T>(props: React.PropsWithChildren<{ data: T, onClick: (item: T) => void, className?: string }>) {
    return <tr className={"listPageTableRow nonHeaderRow"+(props.className?" "+props.className : "")} onClick={() => {
        props.onClick && props.onClick(props.data)
    }}>{props.children}</tr>
}

// TODO: MOVE!
export interface ListTableColumn<T> {
    key: string
    f: (v:T)=>string
}
// TODO: MOVE!
export function NewColumn<T>(key:string,f:(v:T)=>any):ListTableColumn<T> {
    return {key:key,f:f}
}

// TODO: MOVE!
export function ListPageTable<T>({data, onClick, cols,className}: {
    data: T[],
    onClick?: (v: T) => void,
    cols: ListTableColumn<T>[],
    className?: string,
}){
    return <table className={"listPageTable"}>
        <tr className={"listPageTableRow headerRow"}>
            {cols.map((col,i)=>{
                return <th key={i} >{col.key}</th>
            })}
        </tr>
        {data.map((item,i) => {
            return <ListPageTableRow className={className} key={i} data={item} onClick={(v)=>{onClick && onClick(v)}}>{/* TODO: ADD EXPANSION???*/}
                {cols.map((col,i)=>{
                    return <td key={i}>{col.f(item)}</td>
                })}
            </ListPageTableRow>
        })}
    </table>
}

// TODO: MOVE!
export function NumberToDateStr(n: number): string {
    const d = new Date(n)
    return (d.getMonth()+1)+"/"+d.getDate()+"/"+d.getFullYear()
}

export function AgarBatchListPageTable({data, onClick}: ListPageItems<AgarBatchData>) {
    const cols: ListTableColumn<AgarBatchData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Color", (v)=>v.color),
        NewColumn("PC Run", (v)=>v.pcRun),
        NewColumn("Agar Recipe", (v)=>v.agarRecipe),
        NewColumn("Last Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    // TODO: expansion for notes????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

// export function AgarBatchListPageTable({data, onClick}: ListPageItems<AgarBatchData>) {
//     return <table className={"listPageTable"}>
//         <tr>
//             <th>{"ID"}</th>
//             <th>{"Color"}</th>
//             <th>{"PC Run"}</th>
//             <th>{"Agar Recipe"}</th>
//             <th>{"Last Updated"}</th>
//         </tr>
//         {data.map((item) => {
//             return <ListPageTableRow data={item} onClick={(v)=>{onClick && onClick(v)}}>
//                 <td>{item._id}</td>
//                 <td>{item.color}</td>
//                 <td>{item.pcRun}</td>
//                 <td>{item.agarRecipe}</td>
//                 <td>{item.lastUpdated}</td>
//             </ListPageTableRow>
//         })}
//         </table>
// }

export function AgarBatchArea({agarBatchId, headerLevel, offset}: {
    agarBatchId?: string,
    headerLevel?: number,
    offset?: number
}) {
    let linkArea: JSX.Element = <div>{"unknown"}</div>
    if (agarBatchId !== undefined) {
        const displayId = agarBatchId
        linkArea = <EntryLink props={{displayedId: displayId, linkId: displayId, entryType: "agarBatch"}}
                              data-cy-id="AgarBatchAreaEntryLink">
            {displayId}
        </EntryLink>
    }
    return <div data-cy-id="AgarBatchAreaWrapper" className={"agarBatchAreaWrapper"}>
        <div data-cy-id="AgarBatchAreaHeader">{"Agar Batch ID: "}</div>
        {linkArea}
    </div>
}

export function AgarColorArea(
    {current, onSelect}: {
        current: AgarColor, // TODO: may need to be initial
        onSelect?: (s: AgarColor) => void
    }) {

    const {isPending, error, data} = useQuery({
        queryKey: ['colorOptions'],
        queryFn: () => getOptionsResponse("colors")
    })
    const selectorArea = () => {
        if (isPending) {
            return <div>{"LOADING COLOR SELECTOR"}</div>
        }
        if (error !== null) {
            return <div>{"ERROR LOADING COLORS: " + error.message}</div>
        }
        return <SelectorFor disabled={onSelect === undefined} data-cy-id={"Color"} options={data || []}
                            initial={current} updateParent={(colStr: string) => {
            onSelect && onSelect(colStr as AgarColor)
        }}/>
    }
    return <div className={"agarColorArea"}>
        <label className={"newAreaLabel"}>{"Color: "}</label>
        {selectorArea()}
    </div>
}