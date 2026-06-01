'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorTriCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubstrateRecipeData, SubstrateRecipeSelectorCloseable} from "@/app/components/substrateRecipeServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubstrateRecipeArea} from "@/app/components/substrateRecipeClient";
import {NewBagForm} from "@/app/components/bagClient";
import {BagData} from "@/app/components/bagServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertSubstrateBatch(input: any): asserts input is SubstrateBatchData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['recipe', 'string'],
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
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
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

export default function SubstrateBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertSubstrateBatch(data)
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: SubstrateBatchData) => {
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const substrateSubmit = () => {
            const body: any = {
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("substrateBatch", initial._id, body, AssertSubstrateBatch, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("substrateBatch", initial._id), {
            //     method: "POST",
            //     headers: clientPostRequestHeaders,
            //     body: JSON.stringify(body)
            // })
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertSubstrateBatch(entry)
            //         updateInitial(entry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const onViewCreators: OnViewCreatorTriCol[] = [
            {
                txt: "Create Bag",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewBagForm substrateBatchIn={data} handlers={{
                        isTopLevel: false, onCreate: (newItem: BagData) => {
                            return onCreate([{
                                typeText: "Bag",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"bag"}/>
                            }], false)
                        },
                    }}/>
                },
            },
            {
                txt: "Create Fruiting Chamber",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewFruitingChamberForm substrateBatchIn={data} handlers={{
                        isTopLevel: false, onCreate: (newItem: FruitingChamberData) => {
                            return onCreate([{
                                typeText: "Fruiting Chamber",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"fruitingChamber"}/>
                            }], false)
                        },
                    }}/>
                },
            }
        ]
        return (
            <DisplayFormWrapper entryType={"substrateBatch"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Substrate Batch"} entryType={"substrateBatch"}/>
                <OnViewCreatorsTriColArea OnViewCreators={onViewCreators} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Creation Date: "} when={initial.creationDate} readonly={true}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SubstrateRecipeArea id={data.recipe} readonly={true}/> {/* TODO: load name for this recipe? */}
                    </FlexedSinglesGroup>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"}
                                        closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    substrateSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Substrate Batch data format incorrect: " + err}</div>
    }
}

export function NewSubstrateBatchForm({handlers, recipe}: { // TODO: likely rework this whole thing
    handlers: NewEntryInput<SubstrateBatchData>,
    recipe?: SubstrateRecipeData
}) {
    // TODO: do we want the formOpen button outside of the component?
    const [formOpen, setFormOpen] = useState(false)
    const [selectedRecipe, setSelectedRecipe] = useState(recipe)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const submit = () => {
        if (selectedRecipe === undefined) {
            setErr("a recipe must be selected")
            return
        }
        const body: any = {
            recipe: selectedRecipe._id,
            notes: notes
        }
        DoCreateRequest("substrateBatch", body, AssertSubstrateBatch, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("substrateBatch"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body)
        // })
        //     .then(HandleJsonResponse)
        //     .then((entry) => {
        //         AssertSubstrateBatch(entry)
        //         handlers.onCreate && handlers.onCreate(entry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    if (!formOpen) {
        return <div>
            <button className={"basicButton"} onClick={() => {
                setFormOpen(true)
            }}>{"Create new batch"}</button>
        </div>
    }
    return (
        <NewEntryFormWrapper entryType={"substrateBatch"}>
            {/* button to open/close the creator*/}
            <button className={"basicButton"} onClick={() => {
                setFormOpen(false)
            }}>{"Close batch creator"}</button>
            <ErrorDisplay err={err}/>
            <TestAndValidate todos={["ENSURE WORKS PROPERLY FOR BOTH EXISTING AND PICKING"]}>
                {recipe === undefined ?
                    <SubstrateRecipeArea txt={"Substrate Recipe: "} readonly={true} id={selectedRecipe?._id}/> :
                    <SubstrateRecipeSelectorCloseable doSelect={setSelectedRecipe}
                                                      allowCreation={handlers.isTopLevel}
                                                      creatorInPage={false/* TODO: false ok?*/}/>}{/* TODO: closeable vs not?*/}
            </TestAndValidate>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>
        </NewEntryFormWrapper>
    )
}


export const SubstrateBatchArea = ({id, headerLevel, txt, readonly, onSelect}: {
    id?: string,
    headerLevel?: number,
    txt?: string,
    readonly: boolean,
    onSelect?: (d?: SubstrateBatchData) => void
}) => {
    // TODO: FIX THIS WHOLE THING! Update id on prop id update! Store id internally!
    const [open, setOpen] = useState(false)
    const [val, setVal] = useState(id)
    useEffect(() => {
        setVal(id)
    }, [id])
    const updateId = (batch?: SubstrateBatchData) => {
        setVal(batch?._id) // TODO: ensure ok
        onSelect && onSelect(batch)
    }
    let linkArea = () => {
        if (!val) {
            return <div>{"unknown"}</div>
        }
        const tempLink = <EntryLink
            props={{displayedId: val, linkId: val, entryType: "substrateBatch"}}>{val}</EntryLink>
        if (readonly) {
            return tempLink
        }
        return <>
            {tempLink}
            <button className={"basicButton"} onClick={() => {
                setOpen(!open)
            }}>{(open ? "Close Selector" : "Select a new substrate batch")}</button>
        </>
    }
    return <div>
        <div>
            {txt ? txt : "Substrate Batch: "}{linkArea()}
        </div>
        {open && <SubstrateBatchSelector doSelect={updateId}/>}{/* TODO: allow create?*/}
    </div>
}

export function SubstrateBatchListPageTable({data, onClick, withLink}: ListPageItems<SubstrateBatchData>) {
    let cols: ListTableColumn<SubstrateBatchData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Recipe", (v) => v.recipe),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubstrateBatchData) => {
            return <EntryLinkWrapper
                props={{linkId: encodeURI(v._id), entryType: "substrateBatch", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function SubstrateBatchSelectorTable({data, onClick}: ListPageItems<SubstrateBatchData>) {
    return <SubstrateBatchListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function SubstrateBatchSelector(
    {
        doSelect,
        allowCreate,
        creatorInPage,
    }: {
        doSelect: (val: SubstrateBatchData | undefined) => void,
        allowCreate?: boolean
        creatorInPage?: boolean,
    }) {
    const table = (items: SubstrateBatchData[]): JSX.Element => {
        return <SubstrateBatchSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"substrateBatch"} entryTypes={"substrateBatches"} doSelect={doSelect}
                                   asserter={AssertSubstrateBatch}
                                   table={table}>
        {allowCreate && (creatorInPage ? <NewSubstrateBatchForm handlers={{onCreate: doSelect, isTopLevel: false}}/> :
            <div>{"LINK TO CREATOR HERE FIXME"}</div>)}
    </ExistingRecentSelector>
}