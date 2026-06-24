'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorTriCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubstrateRecipeData, SubstrateRecipeSelectorCloseable} from "@/app/components/substrateRecipeServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
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
    OptionalArrayOfType, RequiredKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {SubstrateRecipeArea} from "@/app/components/substrateRecipeClient";
import {NewBagForm} from "@/app/components/bagClient";
import {BagData} from "@/app/components/bagServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
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
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['recipe', 'string'],
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
            throw new Error('Substrate Batch assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return


}

export default function SubstrateBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SubstrateBatchData>) {
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: SubstrateBatchData) => {
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const substrateSubmit = () => {
            const body: any = {
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("substrateBatch", initial._id, body, AssertSubstrateBatch, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SubstrateBatchData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorTriCol[] = [
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
                <ErrorDisplay err={err}/>
                <ID props={{id:data._id, txt:"Substrate Batch", entryType:"substrateBatch"}}/>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
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
                    <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    substrateSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewSubstrateBatchForm({handlers, recipe}: { // TODO: likely rework this whole thing
    handlers: NewEntryInput<SubstrateBatchData>,
    recipe?: SubstrateRecipeData
}) {
    // TODO: do we want the formOpen button outside of the component?
    const [selectedRecipe, setSelectedRecipe] = useState(recipe)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const submit = () => {
        if (selectedRecipe === undefined) {
            setErr("a recipe must be selected")
            return
        }
        const body: any = {
            recipe: selectedRecipe._id,
            notes: notes,
            // Initially created with readonly by all and write by owner
        }
        DoCreateRequest("substrateBatch", body, AssertSubstrateBatch, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new SubstrateBatchData(v)) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return (
        <NewEntryFormWrapper entryType={"substrateBatch"}>
            {/* button to open/close the creator*/}
            <ErrorDisplay err={err}/>
            <TestAndValidate todos={["ENSURE WORKS PROPERLY FOR BOTH EXISTING AND PICKING"]}>
                {recipe === undefined ?
                    <SubstrateRecipeArea txt={"Substrate Recipe: "} readonly={true} id={selectedRecipe?._id}/> :
                    <SubstrateRecipeSelectorCloseable doSelect={setSelectedRecipe}
                                                      allowCreation={handlers.isTopLevel}
                                                      creatorInPage={false/* TODO: false ok?*/}/>}
            </TestAndValidate>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Create new batch"}</button>
        </NewEntryFormWrapper>
    )
}


export const SubstrateBatchArea = ({id, txt}: {
    id?: string,
    txt?: string,
}) => {
    return <div>
        <div className={"inlineChildren"}>
            <div>{txt ? txt : "Substrate Batch: "}</div>
            {id ? <div>
                <EntryLinkForId
                    props={{displayId: id, linkId: id, entryType: "substrateBatch", openInNewTab: false/* TODO: ok?*/}}/>
            </div> :
                <div>{"unknown"}</div>}
        </div>
    </div>
}

export function SubstrateBatchListPageTable({data, onClick, withLink}: ListPageItems<SubstrateBatchData>) {
    let cols: ListTableColumn<SubstrateBatchData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Recipe", (v) => v.recipe, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubstrateBatchData) => {
            return <EntryLinkWrapper
                props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SubstrateBatchData(v)}}/>
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