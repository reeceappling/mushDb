'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    ListResult,
    OnViewCreatorTriCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    CreateNewEntryButton,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey
} from "@/app/components/common";
import {AliasesArea, ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {NewSubstrateBatchForm} from "@/app/components/substrateBatchClient";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {
    DisplayFormWrapper,
    NewEntryFormWrapper
} from "@/app/components/lcRecipeClient";
import {ExistingDualSelector, InlineEntry} from "./agarRecipeClient";

export function AssertSubstrateRecipe(input: any): asserts input is SubstrateRecipeData {
    if (typeof input !== 'object') {
        console.error('Input is not an object! Input is ' + typeof input)
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            console.error('SubRec assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
            throw new Error('SubRec assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            console.error('SubRec assertion failure: optional key ' + key + ' was not valid');
            throw new Error('SubRec assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['aliases', IsString],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('SubRec assertion failure: optional array key ' + key + ' was not valid');
            throw new Error('SubRec assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return


}

// TODO: MOVE!
export function CreatedLinkFor({linkId, typ, linkText}: { linkId: string, typ: string, linkText?: string }) {
    return <EntryLink props={{displayedId: linkText || linkId, linkId: linkId, entryType: typ}}>
        <div>{linkText}</div>
    </EntryLink>
}

export default function SubstrateRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertSubstrateRecipe(data)
        const [initial, setInitial] = useState(data)

        const [name, setName] = useState(initial.name)
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [aliases, setAliases] = useState(initial.aliases || [])
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: SubstrateRecipeData) => {
            setInitial(updated)
            setName(updated.name)
            setIsStandard(updated.standard)
            setAliases(updated.aliases || [])
            setNotes(InitialNotesState(updated.notes))
            setAcl(initial.acl)
        }
        const substrateSubmit = () => {
            fetch(BaseExternalUrl + "/db/update/substrateRecipe/" + initial._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    name: name,
                    standard: isStandard,
                    aliases: aliases,
                    notes: notes,
                    acl: MarshalAcl(acl),
                })
            })
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSubstrateRecipe(entry)
                    updateInitial(entry)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        const ovcs: OnViewCreatorTriCol[] = [
            {
                txt: "Create Substrate Batch",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    // TODO: validate ok
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
                // TODO: any others?
            },
        ]
        return (
            <DisplayFormWrapper entryType={"substrateRecipe"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Substrate Recipe"} entryType={"substrateRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/*TODO: CONSIDER MOVING THIS!*/}
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

                <AliasesArea aliases={aliases} readonly={false} updateParent={setAliases} headerLevel={headerLevel}/>{/* TODO: if empty do not display*/}
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
        return <div>{"ERROR: Substrate Recipe data format incorrect: " + err}</div>
    }
}

export function NewSubstrateRecipeForm({handlers}: { handlers: NewEntryInput<SubstrateRecipeData> }) {
    const [name, setName] = useState("")
    const [aliases, setAliases] = useState<string[]>([])
    const [isStandard, setIsStandard] = useState(false)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    // TODO: TEMPLATE!!!!
    // TODO: handle isTopLevel
    const submit = () => {
        fetch(BaseExternalUrl + "/create/substrateRecipe", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
            body: JSON.stringify({
                name: name,
                aliases: aliases,
                standard: isStandard,
                notes: notes
            })
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertSubstrateRecipe(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return (
        <NewEntryFormWrapper entryType={"substrateRecipe"}>
            <ErrorDisplay err={err}/>
            <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} readonly={false} headerTxt={"Substrate Name: "}/>
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

// export function SubstrateRecipeInline({
//                                           data,
//                                           expandByDefault,
//                                           onClick,
//                                           showMainPageButton,
//                                           idIsLink
//                                       }: InlineProps<SubstrateRecipeData>) {
//     const [expanded, setExpanded] = useState(expandByDefault)
//     const b58id = data._id
//     return <InlineEntry onClick={onClick}>
//         {/* TODO: CHANGE ID TO BUTTON IN CERTAIN SITUATIONS!*/}
//         <InlineSubArea props={{}}>
//             <ID id={b58id} txt={"Substrate Recipe"} entryType={"substrateRecipe"} allowOpenMainPage={showMainPageButton}
//                 linkPage={idIsLink}/>
//             <NameArea currentName={data.name} readonly={true} headerTxt={"Recipe Name: "}/>
//             <AliasesArea readonly={true} aliases={data.aliases}/>
//             <StandardArea isStandard={data.standard} readonly={true}/>
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             <NotesAreaInline notes={data.notes} offset={-1}/>
//             <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                                      expanded={expanded}/>
//     </InlineEntry>
// }

export const SubstrateRecipeArea = ({id, headerLevel, txt, readonly, onSelect}: {
    id?: string,
    headerLevel?: number,
    txt?: string,
    readonly: boolean,
    onSelect?: (d?: SubstrateRecipeData) => void
}) => {
    const [open, setOpen] = useState(false)
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (id !== undefined) {
        const b58id = id
        linkArea =
            <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "substrateRecipe"}}>{b58id}</EntryLink>
        {
            (!readonly && !open) && <button className={"basicButton"} onClick={() => {
                setOpen(true)
            }}>{"Select a new substrate"}</button>
        }
        {
            (!readonly && open) && <div>
                <div>
                    <button className={"basicButton"} onClick={() => {
                        setOpen(true)
                    }}>{"Close Selector"}</button>
                </div>
                <SubstrateRecipeSelector doSelect={r => { // TODO: FIX
                    onSelect && onSelect(r)
                }}/> {/* TODO: allow create? */}
            </div>
        }
    }
    return <div>
        {txt ? txt : "Substrate Recipe: "}{linkArea}
    </div>
}

export function AssertDualListResult<T>(input: any, validateEntry: (inp: any) => void): asserts input is ListResult<T> {
    if (typeof input !== 'object') {
        console.error('Input is not an object! Input is ' + typeof input)
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['recent', validatorForAssertion(validateEntry)], // TODO: ensure ok
        ['standard', validatorForAssertion(validateEntry)],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('optional array key ' + key + ' was not valid')
            throw new Error('optional array key ' + key + ' was not valid');
        }
    }
    return
}

// TODO: move
export function AssertSubRecipeListResult(input: any): asserts input is ListResult<SubstrateRecipeData> {
    AssertDualListResult<SubstrateRecipeData>(input, AssertSubstrateRecipe)
}

export function validatorForAssertion(asserter: ((input: any) => void)) {
    return (inp: any) => {
        try {
            asserter(inp)
            return true
        } catch (e) {
            console.error("error in validatorForAssertion: ", e)
            return false
        }
    }
}

export function SubstrateRecipeListPageTable({data, onClick, withLink}: ListPageItems<SubstrateRecipeData>) {
    let cols: ListTableColumn<SubstrateRecipeData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Name", (v)=>v.name), // TODO: shortname?
        NewColumn("Last Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        })
        // TODO: bonus area for notes??? aliases?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SubstrateRecipeData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"substrateRecipe",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-m"} cols={cols} data={data} onClick={onClick}/>
}
export function SubstrateRecipeSelectorTable({data, onClick}: ListPageItems<SubstrateRecipeData>) {
    return <SubstrateRecipeListPageTable data={data} onClick={onClick} withLink={true} />
}

export function SubstrateRecipeSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate,
        creatorInPage
    }: {
        doSelect: (val: SubstrateRecipeData | undefined) => void,
        allowCreate?: boolean
        creatorInPage?: boolean
    }) {
    const table = (items: SubstrateRecipeData[]):JSX.Element=>{
        return <SubstrateRecipeSelectorTable data={items} onClick={doSelect}/>
    }
    const creator = ()=>{
        if (creatorInPage) {
            return <NewSubstrateRecipeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>
        }
        return <div>{"LINK TO CREATOR HERE FIXME"}</div>
    }
    return <ExistingDualSelector entryType={"substrateRecipe"} entryTypes={"substrateRecipes"} doSelect={doSelect} asserter={AssertSubstrateRecipe}
                                   table={table}>
        {allowCreate && <>{creator()}</>}
    </ExistingDualSelector>
}
