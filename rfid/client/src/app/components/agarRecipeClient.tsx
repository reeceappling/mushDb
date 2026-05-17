'use client'

import React, {JSX, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    Data,
    ListResult,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import LiquidsArea, {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroup,
    LiquidEntriesGroupForNew,
    LiquidsList
} from "@/app/components/formSubcomponents/liquids";
import NutrientsArea, {
    IsValidNutrient,
    Nutrient,
    NutrientEntriesGroup,
    NutrientsEntriesGroupForNew,
    NutrientsList
} from "@/app/components/formSubcomponents/nutrients";
import SugarsArea, {
    IsValidSugar,
    Sugar,
    SugarEntriesGroup,
    SugarEntriesGroupForNew,
    SugarsList
} from "@/app/components/formSubcomponents/sugars";
import {
    AssertArrayResult,
    CreateNewEntryButton,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    IsString,
    ListItemsRequest,
    ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    RequiredArrayOfType,
    ViewInNewTabButton
} from "@/app/components/common";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ErrorDisplay, InlineTitle, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import AdditivesArea, {
    Additive,
    AdditiveEntriesGroup,
    AdditiveEntriesGroupForNew,
    AdditivesList,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {
    Antibiotic,
    AntibioticEntriesGroup,
    AntibioticEntriesGroupForNew,
    AntibioticsDisplay,
    AntibioticsList,
} from "@/app/components/formSubcomponents/antibiotic";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {AssertDualListResult, CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup,
    ListPageTable,
    ListTableColumn,
    NewAgarBatchForm,
    NewColumn,
    NotesFormArea,
    NumberToDateStr
} from "@/app/components/agarBatchClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {DisplayFormWrapper, NewEntryFormWrapper, Subform} from "@/app/components/lcRecipeClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {InputNumber} from "@/app/components/formSubcomponents/numericInput";

export function AssertAgarRecipe(input: any): asserts input is AgarRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['agar', 'number'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            console.error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex optional simple keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            console.error('AgarRecipe assertion failure: optional key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['antibiotics', IsString], // TODO: IsValidAntibiotic
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

// TODO: MOVE
export function dataFor<Type>(vals?: Type[]): Data<Type>[] {
    return (vals || []).map((l) => {
        return {data: l, disabled: false}
    })
}

export default function AgarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }:
    DisplayInput
) {
    try {
        AssertAgarRecipe(data)
        const [initial, setInitial] = useState(data)
        // Required
        const [name, setName] = useState(data.name)
        // Optional
        const [isStandard, setIsStandard] = useState(data.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFor(data.notes || []), new: []})
        const [acl, setAcl] = useState<ACL | undefined>(data.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: AgarRecipeData) => {
            setInitial(updated)
            setName(updated.name)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }

        const agarRecipeSubmit = () => {
            if (name === undefined || name === "") {
                setErr("Name field must not be empty")
                return
            }

            fetch(BaseExternalUrl + "/db/update/agarRecipe/" + initial._id, {
                method: 'Post',
                body: JSON.stringify({
                    name: name,
                    standard: isStandard,
                    notes: notes,
                    acl: MarshalAcl(acl), // TODO; use this everywhere if it works
                }),
                headers: {
                    credentials: 'include',
                    'Cookie': cookies,
                    'Content-type': "application/json"
                },
            })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    try {
                        AssertAgarRecipe(newEntry)
                        updateInitial(newEntry)
                    } catch (er) {
                        throw new Error("failed to decode response:" + JSON.stringify(er))
                    }
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "Create Batch From Recipe", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewAgarBatchForm agarRecipeIn={data} handlers={{
                        onCreate: (newItem: AgarBatchData) => {
                            return onCreate([{
                                typeText: "Agar Batch",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"agarBatch"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            }
        ]
        return (
            <DisplayFormWrapper entryType={"agarRecipe"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <TestAndValidate todos={["Put name at top????"]}>
                    <ID id={data._id} txt={"Agar Recipe"} entryType={"agarRecipe"}/>
                </TestAndValidate>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <NameArea currentName={name} setName={setName}
                                  readonly={readonly}/>{/*TODO: Allow changing??? Make this area longer!*/}
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>{/* TODO: upon change (only when clicked first), deletes users from ACL. FIX THAT*/}
                        <div className={"inlineChildren"}>
                            <div>{"Agar g/L: "}</div>
                            <div>{initial.agar}</div>
                        </div>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>


                <LiquidsArea initialValues={dataFor(initial.liquids)}
                             readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <NutrientsArea initialValues={dataFor(initial.nutrients)}
                               readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <SugarsArea initialValues={dataFor(initial.sugars)} readonly={true}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <AdditivesArea readonly={true}
                               initialValues={dataFor(initial.additives)}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <AntibioticsDisplay antibiotics={initial.antibiotics}/>{/* TODO: FIX AND REFORMAT THIS*/}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {/* TODO: fix the ADD A PROJECT area*/}
                {readonly ? null :
                    <button className={"bottomButton greenButton"} onClick={(e) => {
                        e.stopPropagation();
                        agarRecipeSubmit()
                    }}>{"Update"}</button>}

            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Agar Recipe data format incorrect: " + err}</div>
    }
}

export function NewAgarRecipeForm({handlers}: { handlers: NewEntryInput<AgarRecipeData> }) {
    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [agar, setAgar] = useState(20)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [antibiotics, setAntibiotics] = useState<Antibiotic[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [agarErr, setAgarErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const newAgarRecipeSubmit = () => {
        if (name === "") {
            setErr("name must not be empty")
            return
        }
        if (liquids.length === 0) {
            setErr("at least one liquid must exist")
            return
        }
        let body: any = {
            name: name,
            standard: isStandard,
            agar: agar,
            liquids: liquids,
            // Optional
            nutrients: nutrients.length !== 0 ? nutrients : undefined,
            sugars: sugars.length !== 0 ? sugars : undefined,
            additives: additives.length !== 0 ? additives : undefined,
            antibiotics: antibiotics.length !== 0 ? antibiotics : undefined,
            notes: notes.length !== 0 ? notes : undefined,
        }
        fetch(BaseExternalUrl + "/db/create/agarRecipe", {
            method: 'Post',
            body: JSON.stringify(body),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        }).then(HandleJsonResponse).then((newRecipe) => {
            try {
                AssertAgarRecipe(newRecipe)
                handlers.onCreate && handlers.onCreate(newRecipe)
            } catch (e) {
                setErr("result was not recipe: " + JSON.stringify(e))
            }
        })
            .catch((er) => {
                setErr(JSON.stringify(er))
            })
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <AgarRecipeSelector doSelect={(rec) => {
                if (rec === undefined) {
                    return
                }
                setName(rec.name)
                setIsStandard(rec.standard)
                setAgar(rec.agar)
                setLiquids(rec.liquids)
                setNutrients(rec.nutrients || [])
                setSugars(rec.sugars || [])
                setAdditives(rec.additives || [])
                setAntibiotics(rec.antibiotics || [])
                // TODO: notes?
                setTemplateSelectorOpen(false)
            }}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"agarRecipe"}>
            <ErrorDisplay err={err}/>
            <div>
                {templateRecipeSelector()}
            </div>
            <NameArea classNames={"inlineChildren"} titleClasses={"mr-2"} currentName={name || ""} setName={setName}
                      headerTxt={"Recipe Name: "} readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={false}
                          headerTxt={"Standard Recipe? "}/>
            <div className={"inlineChildren my-4"}>
                <InlineTitle title={"Agar g/L: "} titleClasses={"mr-2"}>
                    <InputNumber readonly={false} value={"" + agar} min={0} max={100}
                                 mode={"integer"} onChange={(e) => {
                        const val = e || ""
                        try {
                            if (val === '' || /^\d*\.?\d*$/.test(val)) {
                                setAgar(parseFloat(val));
                            }
                        } catch {
                            setAgarErr("invalid agar amount")
                        }
                    }} placeholder={"10"} errorMessage={agarErr}/>
                    <div className={"ml-2"}>
                        {agarPer400mL(agar)}
                    </div>
                </InlineTitle>
            </div>
            <div>
                <div>{"Liquids: "}</div>
                <LiquidEntriesGroupForNew currentEntries={liquids} updateParent={setLiquids}/>
            </div>
            <div>
                <div>{"Nutrients: "}</div>
                <NutrientsEntriesGroupForNew currentEntries={nutrients}
                                             updateParent={setNutrients}/>
            </div>
            <div>
                <div>{"Sugars: "}</div>
                <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
            </div>
            <div>
                <div>{"Additives: "}</div>
                <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
            </div>
            <div>
                <div>{"Antibiotics: "}</div>
                <AntibioticEntriesGroupForNew currentEntries={antibiotics}
                                              updateParent={setAntibiotics}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={newAgarRecipeSubmit}/>
        </NewEntryFormWrapper>
    )
}

export function agarPer400mL(agar: number) {
    return <div>{"(" + (agar * 2.0 / 5.0) + " g/400mL)"}</div>
}

export function InlineEntry(props: React.PropsWithChildren<{ onClick?: () => void }>) { // TODO: ADD THIS TO ALL INLINES!!!!!
    return <div className={"inlineEntry"} onClick={(e) => {
        e.stopPropagation()
        props.onClick && props.onClick()
    }}>
        {props.children}
    </div>
    // TODO: add depth to each of these?
}

export function AgarRecipeInline({
                                     data,
                                     expandByDefault,
                                     onClick,
                                     showMainPageButton,
                                     idIsLink
                                 }: InlineProps<AgarRecipeData>) {
    // TODO: do inlines need depth providers?
    const [expanded, setExpanded] = useState(expandByDefault)
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={data._id} txt={"Agar Recipe"} entryType={"agarRecipe"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <NameArea currentName={data.name} readonly={true} headerTxt={"Recipe Name: "}/>
            <StandardArea isStandard={data.standard} readonly={true}/>
            <div>
                <div>{"Agar: " + data.agar + " g/L (" + agarPer400mL(data.agar) + " g/400mL)"}</div>
            </div>
            <LiquidEntriesGroup preexisting={true} readonly={true} initialEntries={data.liquids.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: Liquids (with more on expand)*/}
            <NutrientEntriesGroup preexisting={true} readonly={true} initialEntries={data.nutrients?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: NUTRIENTS (with more on expand)*/}
            <SugarEntriesGroup preexisting={true} readonly={true} initialEntries={data.sugars?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: SUGARS (with more on expand) */}
            <AdditiveEntriesGroup preexisting={true} readonly={true} initialEntries={data.additives?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: ADDITIVES (with more on expand) */}
            <AntibioticEntriesGroup preexisting={true} readonly={true} initialEntries={data.antibiotics?.map((l) => {
                return {data: l, disabled: false}
            })} updateParent={() => {
            }}/>{/* TODO: ANTIBIOTICS (with more on expand) */}

        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes} offset={-1} header={"Notes: "}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

export const AgarRecipeArea = ({agarRecipeBinId}: { agarRecipeBinId?: string }) => {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (agarRecipeBinId !== undefined) {
        const displayId = agarRecipeBinId
        linkArea = <EntryLink
            props={{displayedId: displayId, linkId: displayId, entryType: "agarRecipe"}}> {/* TODO: DISPLAY NAME? */}
            <div>{displayId}</div>
            {/* TODO: NAME? */}
        </EntryLink>
    }
    return <div className={"agarRecipeArea"}>
        <div>{"Agar Recipe ID: "}</div>
        <div>{linkArea}</div>
    </div>
}

// TODO: MOVE!
export function ExistingDualSelector<T>(props: React.PropsWithChildren<{
    doSelect: (val?: T) => void,
    table: (items: T[],onSelect: (v?: T)=>void) => JSX.Element,
    entryType:string,
    entryTypes:string,
    asserter: (val: any)=>void
}>){
    const [err, setErr] = useState<string | undefined>(undefined)
    const [loaded, setLoaded] = React.useState(false);
    const [data, setData] = React.useState<ListResult<T> | undefined>(undefined);
    useEffect(()=>{ListItemsRequest(props.entryTypes).then((result) => {
        try {
            AssertDualListResult<T>(result, props.asserter) // TODO: unnecessary?
            setData(result)
            setLoaded(true)
            return
        } catch (e) {
            console.error(e)
            throw e
        }
    }).catch(e => {
        console.error(e)
        setErr("error on listItems request: " + JSON.stringify(e))
    })},[])
    if (!loaded || data === undefined) {
        return <div>
            <ErrorDisplay err={err}/>
            <div>{"Loading "+props.entryType+" Selector"}</div>
        </div>
    }
    return <Subform>
        <ErrorDisplay err={err}/>
        <SelectorTableWithHeader header={"Recent"} data={data?.recent} onSelect={props.doSelect} table={props.table}/>
        <SelectorTableWithHeader header={"Standard"} data={data?.standard} onSelect={props.doSelect} table={props.table}/>
        <SelectorCreationArea>{props.children}</SelectorCreationArea>
    </Subform>
}
// TODO: move!
export function SelectorCreationArea(props:React.PropsWithChildren<{}>){
    const [creatorOpen, setCreatorOpen] = React.useState(false);
    if (!props.children){
        return null
    }
    if (!creatorOpen){
        return <button className={"buttonFullWidth basicButtonSmall"} onClick={e=>{
            e.stopPropagation();
            setCreatorOpen(true);
        }}>{"Create one instead"}</button>
    }
    return <><button className={"basicButtonSmall"} onClick={e=>{
        e.stopPropagation();
        setCreatorOpen(false);
    }}>{"Close creator"}</button>
        {props.children}
    </>
}
// TODO: MOVE!
export function ExistingRecentSelector<T>(props: React.PropsWithChildren<{
    doSelect: (val?: T) => void,
    table: (items: T[],onSelect: (v?: T)=>void) => JSX.Element,
    entryType:string,
    entryTypes:string,
    asserter: (val: any)=>void
}>){
    const [err, setErr] = useState<string | undefined>(undefined)
    const [loaded, setLoaded] = React.useState(false);
    const [data, setData] = React.useState<T[] | undefined>(undefined);
    useEffect(()=>{ListItemsRequest(props.entryTypes).then((result) => {
        try {
            AssertArrayResult<T>(result, props.asserter) // TODO: unnecessary?
            setLoaded(true)
            setData(result)
        } catch (e) {
            throw e
        }
    }).catch(e => {
        setErr("error on listItems request: " + JSON.stringify(e))
    })},[])
    if (!loaded || data === undefined) {
        return <div>
            <ErrorDisplay err={err}/>
            <div>{"Loading "+props.entryType+" Selector"}</div>
        </div>
    }
    return <Subform>
        <ErrorDisplay err={err}/>
        <SelectorTableWithHeader header={"Recent"} data={data} onSelect={props.doSelect} table={props.table}/>
        <SelectorCreationArea>{props.children}</SelectorCreationArea>
    </Subform>
}

export function SelectorTableWithHeader<T>({header, data,table,onSelect}:{
    header: string,
    data?: T[],
    onSelect: (val?: T) => void,
    table:(items: T[], onSelect: (v?: T) => void)=>JSX.Element
}){
    if(!data || data.length===0){
        return null
    }
    return <>
        <div className={"text-xl"}>{header}</div>
        {table(data,onSelect)}
    </>
}


export function AgarRecipeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: AgarRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: AgarRecipeData[]):JSX.Element=>{
        return <AgarRecipeSelectorTable data={items} onClick={doSelect}
                                          withLink={true/* TODO: recipe id???*/}/>
    }

    return <ExistingDualSelector entryType={"agarRecipe"} entryTypes={"agarRecipes"} doSelect={doSelect} asserter={AssertAgarRecipe}
                                 table={table}>
        {allowCreate && <NewAgarRecipeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingDualSelector>
}

export function StandardAgarRecipeSelector(
    {
        doSelect
    }: {
        doSelect: (val: AgarRecipeData) => void
    }) { // TODO: THIS WHOLE PART!!!
    const [options, setOptions] = useState<AgarRecipeData[]>([])
    // TODO: do selectors need depth providers?
    // TODO: FIX //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);

    // TODO: REMOVE testRecipes for real later!
    const testRecipes: AgarRecipeData[] = [{
        _id: "(agarRecipeId1)",
        name: "(agarRecipeName1)",
        liquids: [{name: LiquidsList[0], pct: 90}, {name: LiquidsList[1], pct: 9}, {name: LiquidsList[2], pct: 1}],
        agar: 20,
        standard: true,
        nutrients: [{nutrient: NutrientsList[0], amount: 2, unit: "pinches"}, {
            nutrient: NutrientsList[1],
            amount: 17,
            unit: "ug"
        }, {nutrient: NutrientsList[2], amount: 3728, unit: "atoms"}],
        sugars: [{type: SugarsList[0], amount: 2, unit: "pinches"}, {
            type: SugarsList[1],
            amount: 17,
            unit: "ug"
        }, {type: SugarsList[2], amount: 3728, unit: "atoms"}],
        additives: [{additive: AdditivesList[0], amount: 2, unit: "pinches"}, {
            additive: AdditivesList[1],
            amount: 17,
            unit: "ug"
        }, {additive: AdditivesList[2], amount: 3728, unit: "atoms"}],
        antibiotics: [AntibioticsList[0], AntibioticsList[1]],
        notes: [{note: "test note 1", time: Date.now()}, {note: "test note 2", time: Date.now()}],
        lastUpdated: Date.now()
    }, {
        _id: "(agarRecipeId2)",
        name: "(agarRecipeName2)",
        liquids: [{name: "water", pct: 100}],
        agar: 20,
        standard: true,
        lastUpdated: Date.now()
    }]
    useEffect(() => {
        setOptions(testRecipes)
        // TODO: REENABLE
        // fetch(BaseExternalUrl+"/get/standardAgarRecipes", { // TODO: ensure correct // TODO: ensure correct (not currently)
        //     method: "GET",
        //     headers: {
        //         credentials: 'include',
        //         // TODO: FIX 'Cookie': cookies,
        //     },
        // })
        //     .then(HandleJsonResponse)
        //     .then((data) => {
        //         setOptions(data as AgarRecipeData[])
        //     })
        //     .catch((error) => {
        //         console.log(error)
        //     }); // TODO: THIS
    }, []); // TODO: what to rerender on?
    // TODO: two sections. Standard, most recent, or create new.
    if (options.length == 0) return <div>
        <div>{"Standard Recipes:"}</div>
        <div>{"Loading Options..."}</div>
    </div>
    return <div>
        <div>{"Standard Recipes:"}</div>
        {options.map((opt, i) => {
            return <AgarRecipeInline data={opt} onClick={() => {
                doSelect(opt)
            }} key={i}/>
        })}
    </div>
}

export function RecentAgarRecipeSelector(
    {
        doSelect
    }: {
        doSelect: (val: AgarRecipeData) => void
    }) {
    // TODO: do slectors need depth providers?
    const [options, setOptions] = useState<AgarRecipeData[]>([])
    // TODO: FIX //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);

    useEffect(() => {
        fetch(BaseExternalUrl + "/get/recentAgarRecipes", { // TODO: ensure correct (not currently)
            method: "GET",
            headers: {
                credentials: 'include',
                // TODO: FIX 'Cookie': cookies,
            },
        })
            .then(HandleJsonResponse)
            .then((data) => {
                setOptions(data as AgarRecipeData[])
            })
            .catch((error) => {
                console.log(error)
            }); // TODO: THIS
    }, [options]);
    if (options.length == 0) return <div>
        <div>{"Recent Recipes:"}</div>
        <div>{"Loading Options..."}</div>
    </div>
    return <div>
        <div>{"Recent Recipes:"}</div>
        {options.map((opt, i) => {
            return <AgarRecipeInline data={opt} onClick={() => {
                doSelect(opt)
            }} key={i}/>
        })}
    </div>
}

// export function AgarRecipeListDisplay({recent, standard, onClick}: TwoListProps<AgarRecipeData>) {
//     const recentArea = () => {
//         if (recent.length === 0) {
//             return null
//         }
//         return <div>
//             {standard.length > 0 && <div>{"Recent Recipes:"}</div>}
//             {recent.map((b, i) => {
//                 return <AgarRecipeInline data={b} onClick={() => {
//                     onClick(b)
//                 }} key={b._id}/>
//             })}
//         </div>
//     }
//     const standardArea = () => {
//         if (standard.length === 0) {
//             return null
//         }
//         return <div>
//             {recent.length > 0 && <div>{"Standard Recipes:"}</div>}
//             {recent.map((b, i) => {
//                 return <AgarRecipeInline data={b} onClick={() => {
//                     onClick(b)
//                 }} key={b._id}/>
//             })}
//         </div>
//     }
//     return <div>
//         {recentArea()}
//         {standardArea()}
//     </div>
// }

export function AgarRecipeListPageTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
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
                    return <div key={v.nutrient + i}>{v.nutrient}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div> // TODO: any more??
                })}
            </div>
        }),
        NewColumn("Antibiotics", (v) => {
            return <div>
                {v.antibiotics && v.antibiotics.map((v, i) => {
                    return <div key={i}>{v}</div>
                })}
            </div>
        }),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })

        // TODO: bonus area for notes???
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "agarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function AgarRecipeSelectorTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })

        // TODO: bonus area for notes???
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <ViewInNewTabButton entryType={"agarRecipe"} id={v._id}/>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}
