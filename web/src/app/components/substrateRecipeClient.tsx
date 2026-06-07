'use client'

import React, {JSX, useContext, useState} from "react";
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
    OptionalArrayOfType, RequiredKey
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {NewSubstrateBatchForm} from "@/app/components/substrateBatchClient";
import TestAndValidate from "@/app/components/testing/untested";
import {
    AclDisplay,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

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
                .then(v=>{
                    updateInitial(new SubstrateRecipeData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorTriCol[] = [
            {
                txt: "Create Substrate Batch",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewSubstrateBatchForm recipe={data} handlers={{
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
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Substrate Recipe"} entryType={"substrateRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <NameArea currentName={name} setName={setName} readonly={readonly} headerLevel={headerLevel}/>

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>

                <AliasesArea aliases={aliases} readonly={false} updateParent={setAliases} headerLevel={headerLevel}/>
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

export function NewSubstrateRecipeForm({handlers}: { handlers: NewEntryInput<SubstrateRecipeData> }) {
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
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return (
        <NewEntryFormWrapper entryType={"substrateRecipe"}>
            <ErrorDisplay err={err}/>
            <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} readonly={false}
                      headerTxt={"Substrate Name: "}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={false}/>
            <TestAndValidate todos={["this whole thing"]}>{/* TODO: ensure NEW is not displayed*/}
                <AliasesArea aliases={aliases} readonly={false} updateParent={setAliases}/>
            </TestAndValidate>
            <NewEntryNotes setNotes={setNotes}/>

            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={submit}/>
        </NewEntryFormWrapper>
    )
}

export function SubstrateRecipeArea({id, headerLevel, txt, readonly, onSelect}: { // TODO: OVERHAUL!!!!
    id?: string,
    headerLevel?: number,
    txt?: string,
    readonly: boolean,
    onSelect?: (d?: SubstrateRecipeData) => void
}){
    const [open, setOpen] = useState(false)
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (id !== undefined) {
        const b58id = id
        linkArea =
            <EntryLinkForId props={{
                openInNewTab: false,
                displayId: b58id,
                linkId: b58id,
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
    </div>
}

export function SubstrateRecipeListPageTable({data, onClick, withLink}: ListPageItems<SubstrateRecipeData>) {
    let cols: ListTableColumn<SubstrateRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
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
    return <ListPageTable className={"text-m"} cols={cols} data={data} onClick={onClick} newClass={v=>{return new SubstrateRecipeData(v)}}/>
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
