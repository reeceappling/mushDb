'use client'

import React, {JSX, useContext, useState} from "react";
import {useQuery,} from '@tanstack/react-query'
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreatedLinkFor,
    dataFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    RequiredKey,
    Subform,
} from "@/app/components/common";
import {AgarRecipeArea,} from "@/app/components/agarRecipeClient";
import {PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {AgarRecipeData, AgarRecipeSelectorCloseable} from "@/app/components/agarRecipeServer";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SelectorFor} from "@/app/components/selector";
import {getOptionsResponse} from "@/app/components/formSubcomponents/server";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl,} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewPlateForm} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {NewSlantForm} from "@/app/components/slantClient";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";

export function AssertAgarBatch(input: any): asserts input is AgarBatchData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['color', 'string'],
        ['pcRun', 'string'],
        ['agarRecipe', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Agar Batch assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('agarBatch assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('agarBatch assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function AgarBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<AgarBatchData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
    const [err, setErr] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL>(initial.acl)
    const updateInitial = (updated: AgarBatchData) => {
        setInitial(updated)
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const agarBatchSubmit = async () => {
        if (notes.new.length === 0 && notes.existing === dataFor(initial.notes)) {
            setErr("No changes found")
            return
        }
        const body: any = {
            notes: notes,
            acl: MarshalAcl(acl),
        }

        DoUpdateRequest("agarBatch", initial._id, body, AssertAgarBatch, allCookies(cookies))
            .then(v => {
                updateInitial(new AgarBatchData(v))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Success",
                        text: "entry updated successfully",
                        isErr: false
                    }})
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Failed",
                        text: "failed to update: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
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
            // TODO: also sticks should be BOILED BEFORE going in the PC! HANDLE BOILING TIME ON STICKS?
            txt: "Create Slants (Before PC)",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <NewSlantForm agarBatchIn={initial} handlers={{
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
        {
            txt: "Add to slant (After PC, irregular)", // TODO: THIS!
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <div>{"AREA NOT IMPLEMENTED YET"}</div>
            },
            needsTesting: true,
        },
    ]
    return (
        <DisplayFormWrapper entryType={"agarBatch"}>
            <ID props={{id:data._id, txt:"Agar Batch", entryType:"agarBatch", linkPage:false, allowOpenMainPage:false}}/>
            <ErrorDisplay data-cy-id={"Error"} err={err}/>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <DateArea data-cy-id={"LastUpdated"} pre={"Last Updated: "} when={initial.lastUpdated}
                          readonly={true}/>
                <div data-cy-id={"Color"}>{"Color: " + data.color}</div>
                <PcRunArea data-cy-id={"Run"} binaryId={initial.pcRun}/>
                <AgarRecipeArea data-cy-id={"Recipe"} agarRecipeBinId={initial.agarRecipe}/>
            </FlexedArea>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>

            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                agarBatchSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function NewAgarBatchForm({handlers, agarRecipeIn, pcRunInp}: {
    handlers: NewEntryInput<AgarBatchData>,
    agarRecipeIn?: AgarRecipeData,
    pcRunInp?: PcRunData
}) {
    const defaultColor = "Clear"
    const {dispatch} = useModalContext();
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunInp)
    const [recipe, setRecipe] = useState<AgarRecipeData | undefined>(agarRecipeIn)
    const [color, setColor] = useState<string>(defaultColor)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
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
        const body: any = {
            color: color,
            pcRun: pcRun._id,
            recipe: recipe._id,
            notes: notes,
        }
        DoCreateRequest("agarBatch", body, AssertAgarBatch, allCookies(cookies))
            .then(v => {
                handlers.onCreate ? handlers.onCreate(new AgarBatchData(v)) : console.warn("no onCreate provided")
            })
            .catch(e => {
                setErr("onCreate handler failed: " + JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"agarBatch"}>
        <div data-cy-id="Header">{"Creating a new agar batch"}</div>
        <ErrorDisplay data-cy-id="Error" err={err}/>
        {pcRunInp ? <PcRunArea binaryId={pcRunInp?._id}/> :
            <Subform>
                <PcRunSelectorCloseable data-cy-id="PcRun" doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                        creatorInPage={handlers.isTopLevel}/>
            </Subform>
        }
        {agarRecipeIn ? <AgarRecipeArea agarRecipe={agarRecipeIn}/> :
            <Subform>
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
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Color", (v) => v.color, true),
        NewColumn("PC Run", (v) => v.pcRun, true),
        NewColumn("Agar Recipe", (v) => v.agarRecipe, true),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit on last?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarBatchData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new AgarBatchData(v)
    }}/>
}

export function AgarBatchArea({agarBatchId, headerLevel, offset}: {
    agarBatchId?: string,
    headerLevel?: number,
    offset?: number
}) {
    return <div data-cy-id="AgarBatchAreaWrapper" className={"agarBatchAreaWrapper"}>
        <div data-cy-id="AgarBatchAreaHeader">{"Agar Batch ID: "}</div>
        {agarBatchId ? <EntryLinkForId props={{linkId: agarBatchId, entryType: "agarBatch"}}
                                       data-cy-id="AgarBatchAreaEntryLink"/> :
            <div>{"unknown"}</div>}
    </div>
}

export function AgarColorArea(
    {initial, onSelect}: {
        initial: string,
        onSelect?: (s: string) => void
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
            onSelect && onSelect(colStr)
        }}/>
    }
    return <div className={"agarColorArea"}>
        <label className={"newAreaLabel"}>{"Color: "}</label>
        {selectorArea()}
    </div>
}

export function AgarBatchSelectorTable({data, onClick, withLink}: ListPageItems<AgarBatchData>) {
    let cols: ListTableColumn<AgarBatchData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Color", (v) => v.color),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarBatchData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new AgarBatchData(v)
    }}/>
}

export function AgarBatchSelector(
    {
        doSelect,
        allowCreate,
    }: {
        doSelect: (val: AgarBatchData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: AgarBatchData[]): JSX.Element => {
        return <AgarBatchSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"agarBatch"} entryTypes={"agarBatches"} doSelect={doSelect}
                                   asserter={AssertAgarBatch}
                                   table={table}>
        {allowCreate && <NewAgarBatchForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
