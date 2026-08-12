'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorTriCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreatedLinkFor,
    CreateNewEntryButton,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingDualSelector,
    FlexedArea,
    FlexedSinglesGroup,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    RequiredKey,
    Subform
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {NewSubstrateBatchForm} from "@/app/components/substrateBatchClient";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {NameModifiable} from "@/app/components/jarRecipeClient";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";

export function AssertSubstrateRecipe(input: any): asserts input is SubstrateRecipeData {
    if (typeof input !== 'object') {
        console.error('Input is not an object! Input is ' + typeof input)
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            console.error('SubRec assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
            throw new Error('SubRec assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Substrate Recipe assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['aliases', IsString],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('SubRec assertion failure: optional array key ' + key + ' was not valid');
            throw new Error('SubRec assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return


}

export default function SubstrateRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SubstrateRecipeData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)

    const [name, setName] = useState(initial.name)
    const [isStandard, setIsStandard] = useState(initial.standard)
    const [aliases, setAliases] = useState(initial.aliases || [])
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
    const [err, setErr] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL>(initial.acl)
    const updateInitial = (updated: SubstrateRecipeData) => {
        setInitial(updated)
        setName(updated.name)
        setIsStandard(updated.standard)
        setAliases(updated.aliases || [])
        setNotes(InitialNotesState(updated.notes))
        setAcl(initial.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const substrateSubmit = () => {
        const body: any = {
            name: name,
            aliases: aliases,
            standard: isStandard,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoUpdateRequest("substrateRecipe", initial._id, body, AssertSubstrateRecipe, allCookies(cookies))
            .then(v => {
                updateInitial(new SubstrateRecipeData(v))
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
    const ovcs: OnViewCreatorTriCol[] = [
        // TODO: bag creation area! (MAYBE MUCH LATER), we probably want to just create bags from the substrateBatch
        // TODO: box creation area! (MAYBE MUCH LATER), we probably want to just create boxes from the substrateBatch and/or jars
        {
            txt: "Create Substrate Batch",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <NewSubstrateBatchForm recipe={data} handlers={{ // TODO: fix so it doesnt have an extra button
                    onCreate: (newItem: SubstrateBatchData) => {
                        return onCreate([{
                            typeText: "Substrate Batch",
                            node: <CreatedLinkFor linkId={newItem._id} typ={"substrateBatch"}/>
                        }], false)
                    },
                    isTopLevel: false,
                }}/>
            },
        },
    ]
    return (
        <DisplayFormWrapper entryType={"substrateRecipe"}>
            <ErrorDisplay err={err}/>
            <ID props={{id: data._id, txt: "Substrate Recipe", entryType: "substrateRecipe"}}>
                <NameModifiable initial={initial.name} readonly={readonly} updateParent={setName}/>
            </ID>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                  headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <AliasesArea initial={initial.aliases} readonly={false} updateParent={setAliases}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"}
                                    closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                substrateSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function NewSubstrateRecipeForm({handlers}: { handlers: NewEntryInput<SubstrateRecipeData> }) {
    const {dispatch} = useModalContext();
    const [name, setName] = useState("")
    const [aliases, setAliases] = useState<string[]>([])
    const [isStandard, setIsStandard] = useState(false)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    // TODO: TEMPLATE!!!!

    const cookies = useContext(CookiesContext)
    const submit = () => {
        // TODO: validate name is valid
        const body = {
            name: name,
            aliases: aliases,
            standard: isStandard,
            notes: notes
            // Initial perms are read by all and write only by owner
        }
        DoCreateRequest("substrateRecipe", body, AssertSubstrateRecipe, allCookies(cookies))
            .then(v => {
                handlers.onCreate ? handlers.onCreate(new SubstrateRecipeData(v)) : console.log("no onCreate provided")
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Success",
                        text: "entry created successfully",
                        isErr: false
                    }})
            })
            .catch(e => {
                setErr(JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: "entry failed to create: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    return (
        <NewEntryFormWrapper entryType={"substrateRecipe"}>
            <ErrorDisplay err={err}/>
            <Subform>
                <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} readonly={false}
                          headerTxt={"Substrate Name: "}/>
                <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={false}/>
            </Subform>
            <Subform>
                <AliasesArea readonly={false} updateParent={setAliases}/>{/* TODO: ensure NEW is not displayed*/}
            </Subform>
            <NewEntryNotes setNotes={setNotes}/>

            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={submit}/>
        </NewEntryFormWrapper>
    )
}

export function SubstrateRecipeArea({id, txt, readonly, onSelect}: { // TODO: OVERHAUL!!!! Allow loading of name!
    id?: string,
    txt?: string,
    readonly: boolean,
    onSelect?: (d?: SubstrateRecipeData) => void
}) {
    const [recipeName, setRecipeName] = useState<string | undefined>(undefined)
    const [recipeId, setRecipeId] = useState<string | undefined>(id)
    useEffect(() => {
        if(id!==undefined && recipeName===undefined){
            // TODO: useEffect on mount if recipeName does not exist but ID does to load the name via the id
        }
    },[])

    const [open, setOpen] = React.useState(false)
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (open) {
        <SubstrateRecipeSelector doSelect={r => {
            if (r !== undefined) {
                onSelect && onSelect(r)
                setRecipeName(r._id)
                setRecipeId(r.name)
                setOpen(false)
            }
        }} allowCreate={false}/>
    }
    if (recipeId !== undefined) {
        // TODO: if name does not exist yet, then create a button to load the name!
        linkArea =
            <EntryLinkForId props={{
                openInNewTab: false,
                displayId: recipeName || recipeId,
                linkId: recipeId,
                entryType: "substrateRecipe"
            }}/>
        //     { // TODO: where does this go?
        //         (!readonly && !open) && <button className={"basicButton"} onClick={() => {
        //             setOpen(true)
        //         }}>{"Select a new substrate"}</button>
        //     }
        //     {
        //         (!readonly && open) && <div>
        //             <div>
        //                 <button className={"basicButton"} onClick={() => {
        //                     setOpen(true)
        //                 }}>{"Close Selector"}</button>
        //             </div>
        //             <SubstrateRecipeSelector doSelect={r => { // TODO: FIX! CLOSEABLE?
        //                 onSelect && onSelect(r)
        //             }}/> {/* TODO: allow create? */}
        //         </div>
        //     }
    }
    return <div>
        {txt ? txt : "Substrate Recipe: "}{linkArea}
        {!readonly && <button className={"basicButtonSmall"} onClick={() => {
            setOpen(true)
        }}>{"Change"}</button>}
    </div>
}

export function SubstrateRecipeListPageTable({data, onClick, withLink}: ListPageItems<SubstrateRecipeData>) {
    let cols: ListTableColumn<SubstrateRecipeData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Name", (v) => v.name, true), // TODO: shortname? or aliases?
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubstrateRecipeData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-m"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new SubstrateRecipeData(v)
    }}/>
}

export function SubstrateRecipeSelectorTable({data, onClick}: ListPageItems<SubstrateRecipeData>) {
    return <SubstrateRecipeListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function SubstrateRecipeSelector(
    {
        doSelect,
        allowCreate,
        creatorInPage
    }: {
        doSelect: (val: SubstrateRecipeData | undefined) => void,
        allowCreate?: boolean
        creatorInPage?: boolean
    }) {
    const table = (items: SubstrateRecipeData[]): JSX.Element => {
        return <SubstrateRecipeSelectorTable data={items} onClick={doSelect}/>
    }
    const creator = () => {
        if (creatorInPage) {
            return <NewSubstrateRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/>
        }
        return <div>{"LINK TO CREATOR HERE FIXME"}</div>
    }
    return <ExistingDualSelector entryType={"substrateRecipe"} entryTypes={"substrateRecipes"} doSelect={doSelect}
                                 asserter={AssertSubstrateRecipe}
                                 table={table}>
        {allowCreate && <>{creator()}</>}
    </ExistingDualSelector>
}
