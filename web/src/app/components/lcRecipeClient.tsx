'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import LiquidsArea, {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroupForNew
} from "@/app/components/formSubcomponents/liquids";
import NutrientsArea, {
    IsValidNutrient,
    Nutrient,
    NutrientsEntriesGroupForNew
} from "@/app/components/formSubcomponents/nutrients";
import SugarsArea, {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew
} from "@/app/components/formSubcomponents/sugars";
import {
    createApiUrlFor,
    CreatedLinkFor,
    dataFor,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingDualSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn,
    NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    RequiredArrayOfType, updateApiUrlFor
} from "@/app/components/common";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import AdditivesArea, {
    Additive,
    AdditiveEntriesGroupForNew,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {AssertLc, NewLcForm} from "@/app/components/lcClient";
import {LcData} from "@/app/components/lcServer";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {AssertJarRecipe} from "@/app/components/jarRecipeClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertLcRecipe(input: any): asserts input is LcRecipeData {
    if (typeof input !== 'object') {
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
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
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

    // complex required array keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid. {' + JSON.stringify(input[key]) + '}{' + JSON.stringify(input) + '}');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function LcRecipeDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {
        AssertLcRecipe(data)
        const [initial, setInitial] = useState(data)

        const [recName, setRecName] = useState(initial.name)
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: LcRecipeData) => {
            setInitial(updated)
            setRecName(updated.name)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const lcRecipeSubmit = () => {
            const body: any = {
                name: recName,
                standard: isStandard,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("lcRecipe",initial._id, body, AssertLcRecipe, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("lcRecipe",initial._id), {
            //     method: "POST",
            //     headers: clientPostRequestHeaders,
            //     body: JSON.stringify({
            //         name: recName,
            //         standard: isStandard,
            //         notes: notes,
            //         acl: MarshalAcl(acl),
            //     })
            // })
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertLcRecipe(entry)
            //         updateInitial(entry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New LC from LcRecipe",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewLcForm lcRecipeIn={initial} handlers={{
                        onCreate: (newItem: LcData) => {
                            return onCreate([{
                                typeText: "Liquid Culture Jar",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"lc"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            }
        ]
        return (
            <DisplayFormWrapper entryType={"lcRecipe"}>
                <ErrorDisplay err={err}/>
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Liquid Culture Recipe"} entryType={"lcRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <TestAndValidate todos={["allow name changes?"]}>
                            <NameArea currentName={data.name} readonly={readonly} headerLevel={headerLevel}
                                      setName={setRecName}/>
                        </TestAndValidate>

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>


                <LiquidsArea initialValues={dataFor(initial.liquids || [])} headerLevel={headerLevel}
                             readonly={true}/>{/* TODO: viewOnlyArea*/}
                <NutrientsArea initialValues={dataFor(initial.nutrients || [])} headerLevel={headerLevel}
                               readonly={true}/>{/* TODO: viewOnlyArea*/}
                <SugarsArea initialValues={dataFor(initial.sugars || [])} headerLevel={headerLevel}
                            readonly={true}/>{/* TODO: viewOnlyArea*/}
                <AdditivesArea initialValues={dataFor(initial.additives || [])} headerLevel={headerLevel}
                               readonly={true}/>{/* TODO: viewOnlyArea*/}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>

                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    lcRecipeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Liquid Culture Recipe data format incorrect: " + err}</div>
    }
}

export function NewLcRecipeForm({handlers}: { handlers: NewEntryInput<LcRecipeData> }) {
    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const loadTemplate = (template: LcRecipeData) => {
        setLiquids(template.liquids)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
    }
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (name === "") {
            setErr("invalid name")
            return
        }
        const body: any = {
            name: name,
            standard: isStandard,
            liquids: liquids,
            nutrients: nutrients,
            sugars: sugars,
            additives: additives,
            notes: notes
        }
        DoCreateRequest("lcRecipe", body, AssertLcRecipe, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("lcRecipe"), {
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify({
        //         name: name,
        //         standard: isStandard,
        //         liquids: liquids,
        //         nutrients: nutrients,
        //         sugars: sugars,
        //         additives: additives,
        //         notes: notes
        //     })
        // })
        //     .then(HandleJsonResponse)
        //     .then((newEntry) => {
        //         AssertLcRecipe(newEntry)
        //         handlers.onCreate && handlers.onCreate(newEntry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <LcRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard
                if (rec === undefined) {
                    return
                }
                loadTemplate(rec)
                setTemplateSelectorOpen(false)
            }} allowCreate={false}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"lcRecipe"}>
            <ErrorDisplay err={err}/>
            <TestAndValidate todos={["TEST THIS"]}>
                {templateRecipeSelector()}
            </TestAndValidate>
            <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} headerTxt={"Recipe name: "}
                      readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} headerTxt={"Standard recipe? "}
                          readonly={false}/>
            <div>{"Liquids"}</div>
            <LiquidEntriesGroupForNew currentEntries={liquids} updateParent={setLiquids}/>
            <div>{"Nutrients"}</div>
            <NutrientsEntriesGroupForNew currentEntries={nutrients} updateParent={setNutrients}/>
            <div>{"Sugars"}</div>
            <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
            <div>{"Additives"}</div>
            <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"greenButton buttonFullWidth"} onClick={createEntry}>{"Create"}</button>
        </NewEntryFormWrapper>
    )
}

export function LcRecipeArea({lcRecipeId, headerLevel, offset}: {
    lcRecipeId?: string,
    headerLevel?: number,
    offset?: number
}) {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (lcRecipeId !== undefined) {
        const b58id = lcRecipeId
        linkArea = <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "lcRecipe"}}>
            <div>{b58id}</div>
        </EntryLink>
    }
    return <div className={"lcRecipeArea"}>
        <div>{"Liquid Culture Recipe ID: "}</div>
        {linkArea}
    </div>
}

export function LcRecipeListPageTable({data, onClick, withLink}: ListPageItems<LcRecipeData>) {
    let cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Liquids", (v) => {
            return <div>
                {v.liquids.map((l, i) => {
                    return <div key={l.name + i}>{l.name}</div>
                })}
            </div>
        }),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div>
                })}
            </div>
        }),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div>
                })}
            </div>
        }),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div>
                })}
            </div>
        }),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "lcRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function LcRecipeSelectorTable({data, onClick}: ListPageItems<LcRecipeData>) {
    let cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "lcRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function LcRecipeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: LcRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: LcRecipeData[]): JSX.Element => {
        return <LcRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"lcRecipe"} entryTypes={"lcRecipes"} doSelect={doSelect}
                                 asserter={AssertLcRecipe}
                                 table={table}>
        {allowCreate && <NewLcRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingDualSelector>
}
