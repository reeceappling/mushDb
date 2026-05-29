'use client'

import React, {JSX, useEffect, useState} from "react";
import {useQuery,} from '@tanstack/react-query'
import NotesArea, {
    IsValidNote,
    Note,
    NotesAreaInline, NotesAreaViewSubcomponent,
    NotesAreaOld,
    NotesGrid, SingleNoteV2, NewEntryNotes, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {AgarBatchData, AgarColor} from "@/app/components/agarBatchServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreatedLinkFor,
    dataFor, DisplayFormWrapper,
    DisplayInput,
    ExistingRecentSelector,
    FlexedArea,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey, Subform,
} from "@/app/components/common";
import {
    AgarRecipeArea,
} from "@/app/components/agarRecipeClient";
import {PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {AgarRecipeData, AgarRecipeSelectorCloseable} from "@/app/components/agarRecipeServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SelectorFor} from "@/app/components/selector";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth,} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewPlateForm} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {NewSlantForm} from "@/app/components/slantClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";

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
                txt: "Create Plates",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewPlateForm agarBatchIn={data} handlers={{
                        onCreate: (newItem: PlateData) => {
                            return onCreate([{
                                typeText: "Plate",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"plate"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
            {
                // TODO: Slants are poured BEFORE PCing! The Agar batch will already have the PC Run on it though, so we shouldnt worry about it.
                // TODO: also sticks should be boiled BEFORE going in the PC!
                txt: "Create Slants",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewSlantForm agarBatchIn={data} handlers={{
                        onCreate: (newItem: PlateData) => {
                            return onCreate([{
                                typeText: "Slant",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"slant"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            },
        ]
        return (
            <DisplayFormWrapper entryType={"agarBatch"}>
                <ID txt={"Agar Batch"} id={data._id} entryType={"agarBatch"} linkPage={false} allowOpenMainPage={false}
                    data-cy-id={"Id"}/>
                <ErrorDisplay data-cy-id={"Error"} err={err} headerLevel={headerLevel}/>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
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
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: AgarBatch data format incorrect: " + err}</div>
    }
}

// TODO: NOT WORKING IN SELECTOR!
export function NewAgarBatchForm({handlers, agarRecipeIn, pcRunInp}: {
    handlers: NewEntryInput<AgarBatchData>,
    agarRecipeIn?: AgarRecipeData,
    pcRunInp?: PcRunData
}) {
    const defaultColor: AgarColor = "Clear"
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunInp)
    const [recipe, setRecipe] = useState<AgarRecipeData | undefined>(agarRecipeIn)
    const [color, setColor] = useState<AgarColor>(defaultColor)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
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
        fetch(BaseExternalUrl + "/db/create/agarBatch", {
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
        {pcRunInp ? <PcRunArea binaryId={pcRunInp?._id}/> :
            <Subform >
                <PcRunSelectorCloseable data-cy-id="PcRun" doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                        creatorInPage={handlers.isTopLevel}/>
            </Subform>
        }
        {agarRecipeIn ? <AgarRecipeArea agarRecipeBinId={agarRecipeIn?._id}/> :
            <Subform >
                <AgarRecipeSelectorCloseable /* TODO: consider using subform on other closeables?*/
                    doSelect={setRecipe}
                    txt={"Agar Recipe: "}
                    allowCreation={true}
                    creatorInPage={false}/>
            </Subform>
        }
        <AgarColorArea data-cy-id={"Color"} initial={defaultColor} onSelect={setColor}/>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"bottomButton greenButton"} onClick={newAgarBatchSubmit}>{"Submit"}</button>
    </NewEntryFormWrapper>
}

export function AgarBatchListPageTable({data, onClick, withLink}: ListPageItems<AgarBatchData>) {
    let cols: ListTableColumn<AgarBatchData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Color", (v)=>v.color),
        NewColumn("PC Run", (v)=>v.pcRun),
        NewColumn("Agar Recipe", (v)=>v.agarRecipe),
        NewColumn("Last Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarBatchData)=>{
            return <EntryLinkWrapper props={{linkId:v._id,entryType:"agarBatch",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

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
    {initial,onSelect}: {
        initial: AgarColor,
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
                            initial={initial} updateParent={(colStr) => {
            onSelect && onSelect(colStr as AgarColor)
        }}/>
    }
    return <div className={"agarColorArea"}>
        <label className={"newAreaLabel"}>{"Color: "}</label>
        {selectorArea()}
    </div>
}

export function AgarBatchSelectorTable({data, onClick, withLink}: ListPageItems<AgarBatchData>) {
    let cols: ListTableColumn<AgarBatchData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Color", (v)=>v.color),
        NewColumn("Last Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarBatchData)=>{
            return <EntryLinkWrapper props={{linkId:v._id,entryType:"agarBatch",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function AgarBatchSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: AgarBatchData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: AgarBatchData[]):JSX.Element=>{
        return <AgarBatchSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"agarBatch"} entryTypes={"agarBatches"} doSelect={doSelect} asserter={AssertAgarBatch}
                                 table={table}>
        {allowCreate && <NewAgarBatchForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
