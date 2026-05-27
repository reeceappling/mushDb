'use client'

import React, {JSX, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorTriCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubstrateRecipeData, TestSubstrateRecipeOk} from "@/app/components/substrateRecipeServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SelectorProps} from "@/app/components/selector";
import Centered from "@/app/components/commonServer";
import {SubstrateBatchData, TestSubstrateBatchOkStd} from "@/app/components/substrateBatchServer";
import {CreatedLinkFor, SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {NewBagForm} from "@/app/components/bagClient";
import {BagData} from "@/app/components/bagServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import TestAndValidate from "@/app/components/testing/untested";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {ExistingRecentSelector, InlineEntry} from "./agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SlantData} from "@/app/components/slantServer";
import {AssertSlant, NewSlantForm} from "@/app/components/slantClient";

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
        id, readonly, data, headerLevel, isTopLevel, cookies
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
        const substrateSubmit = () => {
            fetch(BaseExternalUrl + "/db/update/substrateBatch/" + initial._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    // TODO: may need 'Cookie': cookies,
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    notes: notes,
                    acl: MarshalAcl(acl),
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSubstrateBatch(entry)
                    updateInitial(entry)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
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
                // TODO: any others?
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
                <OnViewCreatorsTriColArea OnViewCreators={onViewCreators} readonly={readonly}/>{/* TODO: MOVE?*/}
                <FlexedArea>
                    <FlexedSinglesGroup>{/*TODO: ALL THESE GROUPS!*/}
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
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
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
    // TODO: isTopLevel handlers
    const submit = () => {
        if (selectedRecipe === undefined) {
            setErr("a recipe must be selected")
            return
        } else {
            fetch(BaseExternalUrl + "/create/substrateBatch", {
                method: "POST",
                headers: {
                    credentials: 'include',
                    // TODO: may need 'Cookie': cookies,
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    recipe: selectedRecipe._id,
                    notes: notes
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSubstrateBatch(entry)
                    handlers.onCreate && handlers.onCreate(entry)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
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
                    <SubstrateRecipeSelector doSelect={setSelectedRecipe}
                                             allowCreate={handlers.isTopLevel}/>}
            </TestAndValidate>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>
        </NewEntryFormWrapper>
    )
}

// export function SubstrateBatchInline({
//                                          data,
//                                          expandByDefault,
//                                          onClick,
//                                          showMainPageButton,
//                                          idIsLink
//                                      }: InlineProps<SubstrateBatchData>) {
//     const [expanded, setExpanded] = useState(expandByDefault)
//     const b58id = data._id
//     return <InlineEntry onClick={onClick}>
//         {/* TODO: CHANGE ID TO BUTTON IN CERTAIN SITUATIONS!*/}
//         <InlineSubArea props={{}}>
//             <ID id={b58id} txt={"Substrate Batch"} entryType={"substrateBatch"} allowOpenMainPage={showMainPageButton}
//                 linkPage={idIsLink}/>
//             <SubstrateRecipeArea id={data.recipe} readonly={true} txt={"Recipe: "}/>
//             <DateArea readonly={true} when={data.creationDate} pre={"Created: "}/>
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             <NotesAreaInline notes={data.notes} offset={-1}/>
//             <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                                      expanded={expanded}/>
//     </InlineEntry>
// }

export const SubstrateBatchArea = ({id, headerLevel, txt, readonly, onSelect}: {
    id?: string,
    headerLevel?: number,
    txt?: string,
    readonly: boolean,
    onSelect?: (d?: SubstrateBatchData) => void
}) => {
    // TODO: does sub batch area need depth incremented?
    const [open, setOpen] = useState(false)
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (id !== undefined) {
        const b58id = id
        linkArea =
            <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "substrateBatch"}}>{b58id}</EntryLink>
        {
            (!readonly && !open) && <button className={"basicButton"} onClick={() => {
                setOpen(true)
            }}>{"Select a new substrate batch"}</button>
        }
        {
            (!readonly && open) && <div>
                <div>
                    <button className={"basicButton"} onClick={() => {
                        setOpen(true)
                    }}>{"Close Selector"}</button>
                </div>
                <SubstrateBatchSelector doSelect={r => { // TODO: FIX
                    onSelect && onSelect(r)
                }}/>{/* TODO: allow create?*/}
            </div>
        }
    }
    return <div>
        {txt ? txt : "Substrate Batch: "}{linkArea}
    </div>
}

export function SubstrateBatchListPageTable({data, onClick, withLink}: ListPageItems<SubstrateBatchData>) {
    let cols: ListTableColumn<SubstrateBatchData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Recipe", (v)=>v.recipe),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubstrateBatchData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"substrateBatch",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function SubstrateBatchSelectorTable({data, onClick}: ListPageItems<SubstrateBatchData>) {
    return <SubstrateBatchListPageTable data={data} onClick={onClick} withLink={true} />
}
export function SubstrateBatchSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate,
        creatorInPage,
    }: {
        doSelect: (val: SubstrateBatchData | undefined) => void,
        allowCreate?: boolean
        creatorInPage?: boolean, // TODO: ADD THIS EVERYWHERE ELSE!!!!!!
    }) {
    const table = (items: SubstrateBatchData[]):JSX.Element=>{
        return <SubstrateBatchSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"substrateBatch"} entryTypes={"substrateBatches"} doSelect={doSelect} asserter={AssertSubstrateBatch}
                                   table={table}>
        {allowCreate && (creatorInPage?<NewSubstrateBatchForm handlers={{onCreate: doSelect,isTopLevel: false}}/>:
            <div>{"LINK TO CREATOR HERE FIXME"}</div>)}
    </ExistingRecentSelector>
}